package josuke

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"sync"
	"time"
)

// queueSize is the number of jobs that can wait for the worker before new ones are rejected.
const queueSize = 100

// job holds the actions triggered by one webhook request.
type job struct {
	id       uint64
	desc     string
	actions  []HookAction
	queuedAt time.Time
	// logger prefixes every line with the job id, so interleaved requests can be told apart.
	logger *log.Logger
	// ctx is cancelled when the job is stopped, which kills its running command.
	ctx    context.Context
	cancel context.CancelFunc

	startedAt time.Time // guarded by jobQueue.mu
}

func newJob(id uint64, desc string, actions []HookAction, queuedAt time.Time) *job {
	ctx, cancel := context.WithCancel(context.Background())
	return &job{
		id:       id,
		desc:     desc,
		actions:  actions,
		queuedAt: queuedAt,
		logger:   log.New(log.Writer(), fmt.Sprintf("job#%d ", id), log.Flags()|log.Lmsgprefix),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// jobQueue runs jobs one at a time, in the order they were pushed.
// Commands share the process working directory and user, and two builds of
// the same project would step on each other, so they must never overlap.
type jobQueue struct {
	run  func(*job)
	size int
	path string // file the queue is saved to, none if empty

	mu      sync.Mutex
	ready   *sync.Cond // signaled when a job is pushed
	waiting []*job
	running *job
	lastID  uint64
}

// newJobQueue creates the queue and starts its worker. If path is set, the queue is
// saved to this file at each change, and the jobs waiting in it are queued again.
func newJobQueue(size int, path string, run func(*job)) (*jobQueue, error) {
	q := &jobQueue{run: run, size: size, path: path}
	q.ready = sync.NewCond(&q.mu)
	if err := q.load(); err != nil {
		return nil, err
	}
	go q.work()
	return q, nil
}

// push queues the actions without blocking. It returns nil if the queue is full.
func (q *jobQueue) push(desc string, actions []HookAction) *job {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.lastID++
	j := newJob(q.lastID, desc, actions, time.Now())
	if len(q.waiting) >= q.size {
		j.cancel()
		j.logger.Printf("[ERR ] rejected %s: queue is full (%d jobs waiting)\n", desc, len(q.waiting))
		return nil
	}
	ahead := len(q.waiting)
	if q.running != nil {
		ahead++
	}
	q.waiting = append(q.waiting, j)
	q.save()
	q.ready.Signal()
	j.logger.Printf("[INFO] queued %s, %d job(s) ahead\n", desc, ahead)
	return j
}

// jobStatus is a snapshot of a job, as listed by the jobs API.
type jobStatus struct {
	ID          uint64     `json:"id"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // running, stopping or waiting
	QueuedAt    time.Time  `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
}

// list returns the running job, if any, followed by the waiting ones in order.
func (q *jobQueue) list() []jobStatus {
	q.mu.Lock()
	defer q.mu.Unlock()

	statuses := make([]jobStatus, 0, len(q.waiting)+1)
	if j := q.running; j != nil {
		status := "running"
		if j.ctx.Err() != nil {
			status = "stopping"
		}
		startedAt := j.startedAt
		statuses = append(statuses, jobStatus{j.id, j.desc, status, j.queuedAt, &startedAt})
	}
	for _, j := range q.waiting {
		statuses = append(statuses, jobStatus{j.id, j.desc, "waiting", j.queuedAt, nil})
	}
	return statuses
}

// stop removes a waiting job from the queue, or kills the command of the running one.
// It returns the job status before it was stopped, false if there is no such job.
func (q *jobQueue) stop(id uint64) (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if j := q.running; j != nil && j.id == id {
		j.logger.Println("[WARN] stopping")
		j.cancel()
		return "running", true
	}
	for i, j := range q.waiting {
		if j.id == id {
			q.waiting = slices.Delete(q.waiting, i, i+1)
			q.save()
			j.cancel()
			j.logger.Println("[WARN] removed from the queue")
			return "waiting", true
		}
	}
	return "", false
}

func (q *jobQueue) work() {
	for {
		q.mu.Lock()
		for len(q.waiting) == 0 {
			q.ready.Wait()
		}
		j := q.waiting[0]
		q.waiting = slices.Delete(q.waiting, 0, 1)
		j.startedAt = time.Now()
		q.running = j
		q.save()
		q.mu.Unlock()

		q.runSafely(j)

		q.mu.Lock()
		q.running = nil
		q.save()
		q.mu.Unlock()
		j.cancel()
	}
}

// runSafely keeps the worker alive if a job panics, otherwise the whole server would go down.
func (q *jobQueue) runSafely(j *job) {
	defer func() {
		if r := recover(); r != nil {
			j.logger.Printf("[ERR ] panic: %v\n%s", r, debug.Stack())
		}
	}()
	q.run(j)
}

// runJob executes the job actions in order. A failing action does not prevent the next ones from running.
func runJob(j *job) {
	j.logger.Printf("[INFO] starting %s (waited %s)\n", j.desc, time.Since(j.queuedAt).Round(time.Second))
	start := time.Now()
	failed := false
	for _, ha := range j.actions {
		if ha.Action == nil {
			continue
		}
		err := ha.Action.execute(j.ctx, ha.Info, j.logger)
		if j.ctx.Err() != nil {
			j.logger.Printf("[WARN] stopped: %v\n", err)
			return
		}
		if err != nil {
			failed = true
			j.logger.Printf("[ERR ] action %s failed: %s\n", ha.Action.Action, err)
		}
	}

	elapsed := time.Since(start).Round(100 * time.Millisecond)
	if failed {
		j.logger.Printf("[ERR ] finished with errors in %s\n", elapsed)
		return
	}
	j.logger.Printf("[INFO] finished successfully in %s\n", elapsed)
}
