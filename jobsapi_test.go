package josuke

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Jobs_API_Lists_And_Stops_Jobs(t *testing.T) {
	started := make(chan uint64, 2)
	j := &Josuke{
		JobsAPI: &JobsAPI{Token: "s3cr3t"},
		queue:   newTestQueue(t, 2, "", blockUntilStopped(started)),
	}
	mux := http.NewServeMux()
	assert.Equal(t, "/jobs", j.registerJobsAPI(mux))
	call := func(method, path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	j.queue.push("first", nil)
	waitFor(t, started)
	j.queue.push("second", nil)

	assert.Equal(t, http.StatusUnauthorized, call("GET", "/jobs", "").Code)
	assert.Equal(t, http.StatusUnauthorized, call("GET", "/jobs", "wrong").Code)
	assert.Equal(t, http.StatusUnauthorized, call("POST", "/jobs/1/stop", "wrong").Code)

	rec := call("GET", "/jobs", "s3cr3t")
	require.Equal(t, http.StatusOK, rec.Code)
	var statuses []jobStatus
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &statuses))
	require.Len(t, statuses, 2)
	assert.Equal(t, "first", statuses[0].Description)
	assert.Equal(t, "running", statuses[0].Status)
	assert.Equal(t, "waiting", statuses[1].Status)

	rec = call("POST", "/jobs/2/stop", "s3cr3t")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id": 2, "status": "removed"}`, rec.Body.String())

	rec = call("POST", "/jobs/1/stop", "s3cr3t")
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.JSONEq(t, `{"id": 1, "status": "stopping"}`, rec.Body.String())
	waitIdle(t, j.queue)

	assert.Equal(t, http.StatusNotFound, call("POST", "/jobs/1/stop", "s3cr3t").Code)
	assert.Equal(t, http.StatusBadRequest, call("POST", "/jobs/abc/stop", "s3cr3t").Code)
	assert.Equal(t, http.StatusMethodNotAllowed, call("GET", "/jobs/1/stop", "s3cr3t").Code)
}
