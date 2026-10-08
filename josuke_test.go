package josuke

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipOnWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs sh")
	}
}

func Test_ExecuteCommand_Logs_Stdout_And_Stderr(t *testing.T) {
	skipOnWindows(t)
	var buf bytes.Buffer

	err := ExecuteCommand(context.Background(), []string{"sh", "-c", "echo out; echo err >&2; printf last"}, &Info{}, log.New(&buf, "", 0))

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "[INFO] sh | out\n[INFO] sh | err\n[INFO] sh | last\n")
	assert.Contains(t, buf.String(), "[INFO] sh done in ")
}

func Test_ExecuteCommand_Returns_Exit_Error_And_Logs_Output(t *testing.T) {
	skipOnWindows(t)
	var buf bytes.Buffer

	err := ExecuteCommand(context.Background(), []string{"sh", "-c", "echo boom >&2; exit 3"}, &Info{}, log.New(&buf, "", 0))

	assert.ErrorContains(t, err, "exit status 3")
	assert.Contains(t, buf.String(), "[INFO] sh | boom\n")
}

func Test_ExecuteCommand_Does_Not_Replace_Placeholders_In_Config(t *testing.T) {
	skipOnWindows(t)
	command := []string{"echo", "%payload_path%"}
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	require.NoError(t, ExecuteCommand(context.Background(), command, &Info{PayloadPath: "/first"}, logger))
	require.NoError(t, ExecuteCommand(context.Background(), command, &Info{PayloadPath: "/second"}, logger))

	assert.Equal(t, []string{"echo", "%payload_path%"}, command)
	assert.Contains(t, buf.String(), "echo | /second\n")
}

func Test_ExecuteCommand_Stops_The_Command_And_Its_Children(t *testing.T) {
	skipOnWindows(t)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(200*time.Millisecond, cancel)
	var buf bytes.Buffer
	start := time.Now()

	// sleep runs in a child of sh: if only sh was killed, sleep would hold the output until waitDelay.
	err := ExecuteCommand(ctx, []string{"sh", "-c", "echo started; sleep 30; echo never"}, &Info{}, log.New(&buf, "", 0))

	assert.ErrorContains(t, err, "interrupted")
	assert.Less(t, time.Since(start), waitDelay/2)
	assert.Contains(t, buf.String(), "sh | started")
	assert.NotContains(t, buf.String(), "sh | never")
}

func Test_Webhook_Request_Queues_A_Job_And_Answers_Right_Away(t *testing.T) {
	queued := make(chan *job, 1)
	j := &Josuke{
		Deployment: []*Repo{{
			Name: "monkeydioude/josuke",
			Branches: []Branch{{
				Name:    "master",
				Actions: []Action{{Action: "push", Commands: [][]string{{"make"}}}},
			}},
		}},
		queue: newTestQueue(t, 1, "", func(jb *job) { queued <- jb }),
	}
	hh, err := NewHookHandler(j, &Hook{Name: "webhook", Type: "webhook"})
	require.NoError(t, err)
	body, err := os.Open("testdata/commit-payload.json")
	require.NoError(t, err)
	defer body.Close()
	req := httptest.NewRequest(http.MethodPost, "/josuke/webhook", body)
	req.Header.Set("x-webhook-event", "push")
	rec := httptest.NewRecorder()

	hh.HookDef.Handler(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, "queued job #1\n", rec.Body.String())
	jb := <-queued
	assert.Equal(t, "webhook push on monkeydioude/josuke@master", jb.desc)
	require.Len(t, jb.actions, 1)
	assert.Equal(t, [][]string{{"make"}}, jb.actions[0].Action.Commands)
}
