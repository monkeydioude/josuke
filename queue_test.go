package josuke

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitFor[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
		panic("unreachable")
	}
}

func newTestQueue(t *testing.T, size int, path string, run func(*job)) *jobQueue {
	t.Helper()
	q, err := newJobQueue(size, path, run)
	require.NoError(t, err)
	return q
}

// waitIdle waits for the queue to be empty, which also means its last save is done.
func waitIdle(t *testing.T, q *jobQueue) {
	t.Helper()
	require.Eventually(t, func() bool { return len(q.list()) == 0 }, 5*time.Second, time.Millisecond)
}

// blockUntilStopped is a job runner reporting the jobs it starts, and running until they are stopped.
func blockUntilStopped(started chan<- uint64) func(*job) {
	return func(j *job) {
		started <- j.id
		<-j.ctx.Done()
	}
}

func Test_Queue_Runs_Jobs_One_At_A_Time_In_Order(t *testing.T) {
	const n = 5
	var mu sync.Mutex
	running, maxRunning := 0, 0
	var order []uint64
	done := make(chan struct{})

	q := newTestQueue(t, n, "", func(j *job) {
		mu.Lock()
		running++
		maxRunning = max(maxRunning, running)
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		running--
		order = append(order, j.id)
		if len(order) == n {
			close(done)
		}
		mu.Unlock()
	})
	for range n {
		require.NotNil(t, q.push("test", nil))
	}

	waitFor(t, done)
	assert.Equal(t, 1, maxRunning)
	assert.Equal(t, []uint64{1, 2, 3, 4, 5}, order)
}

func Test_Queue_Rejects_Jobs_When_Full(t *testing.T) {
	started := make(chan uint64, 1)
	q := newTestQueue(t, 1, "", blockUntilStopped(started))

	require.NotNil(t, q.push("running", nil))
	waitFor(t, started)
	require.NotNil(t, q.push("waiting", nil))
	assert.Nil(t, q.push("rejected", nil))

	q.stop(2)
	q.stop(1)
}

func Test_Queue_Keeps_Running_After_A_Panic(t *testing.T) {
	done := make(chan struct{})
	q := newTestQueue(t, 2, "", func(j *job) {
		if j.id == 1 {
			panic("boom")
		}
		close(done)
	})
	q.push("panics", nil)
	q.push("runs", nil)

	waitFor(t, done)
}

func Test_Queue_Lists_And_Stops_Jobs(t *testing.T) {
	started := make(chan uint64, 3)
	q := newTestQueue(t, 3, "", blockUntilStopped(started))
	q.push("first", nil)
	waitFor(t, started)
	q.push("second", nil)
	q.push("third", nil)

	statuses := q.list()
	require.Len(t, statuses, 3)
	assert.Equal(t, []uint64{1, 2, 3}, []uint64{statuses[0].ID, statuses[1].ID, statuses[2].ID})
	assert.Equal(t, []string{"running", "waiting", "waiting"}, []string{statuses[0].Status, statuses[1].Status, statuses[2].Status})
	assert.NotNil(t, statuses[0].StartedAt)
	assert.Nil(t, statuses[1].StartedAt)

	status, ok := q.stop(2)
	assert.True(t, ok)
	assert.Equal(t, "waiting", status)
	status, ok = q.stop(1)
	assert.True(t, ok)
	assert.Equal(t, "running", status)
	_, ok = q.stop(42)
	assert.False(t, ok)

	// The second job was removed: the third one runs right after the first one.
	assert.Equal(t, uint64(3), waitFor(t, started))
	q.stop(3)
	waitIdle(t, q)
}

func Test_Queue_Saves_Itself_To_Its_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	started := make(chan uint64, 2)
	q := newTestQueue(t, 2, path, blockUntilStopped(started))
	actions := []HookAction{{Action: &Action{Action: "push", Commands: [][]string{{"make"}}}, Info: &Info{BaseDir: "/var/www"}}}

	q.push("first", nil)
	waitFor(t, started)
	q.push("second", actions)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var saved savedQueue
	require.NoError(t, json.Unmarshal(data, &saved))
	assert.Equal(t, uint64(2), saved.LastID)
	require.NotNil(t, saved.Running)
	assert.Equal(t, uint64(1), saved.Running.ID)
	require.Len(t, saved.Waiting, 1)
	assert.Equal(t, "second", saved.Waiting[0].Description)
	assert.Equal(t, actions, saved.Waiting[0].Actions)

	q.stop(1)
	waitFor(t, started)
	q.stop(2)
	waitIdle(t, q)
}

func Test_Queue_Restores_Waiting_Jobs_From_Its_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	data, err := json.Marshal(savedQueue{
		LastID:  3,
		Running: &savedJob{ID: 1, Description: "interrupted"},
		Waiting: []savedJob{{ID: 2, Description: "second"}, {ID: 3, Description: "third"}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0600))

	ran := make(chan uint64, 3)
	q := newTestQueue(t, 3, path, func(j *job) { ran <- j.id })

	// The interrupted job is not run again.
	assert.Equal(t, uint64(2), waitFor(t, ran))
	assert.Equal(t, uint64(3), waitFor(t, ran))
	assert.Equal(t, uint64(4), q.push("new", nil).id)
	assert.Equal(t, uint64(4), waitFor(t, ran))
	waitIdle(t, q)

	data, err = os.ReadFile(path)
	require.NoError(t, err)
	var saved savedQueue
	require.NoError(t, json.Unmarshal(data, &saved))
	assert.Equal(t, savedQueue{LastID: 4, Waiting: []savedJob{}}, saved)
}

func Test_Queue_Fails_On_A_Corrupted_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	require.NoError(t, os.WriteFile(path, []byte("{nope"), 0600))

	_, err := newJobQueue(1, path, runJob)

	assert.ErrorContains(t, err, "remove it to start with an empty queue")
}
