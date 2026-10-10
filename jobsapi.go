package josuke

import (
	"cmp"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// JobsAPI configures the HTTP API listing and stopping jobs.
type JobsAPI struct {
	// Route prefix of the API, defaults to /jobs.
	Route string `json:"route,omitempty" yaml:"route,omitempty"`
	// Token expected in the request Authorization header, as "Bearer <token>". Required.
	Token string `json:"token" yaml:"token"`
}

// HandleJobs declares the HTTP handlers of the jobs API, if configured:
//   - GET <route> lists the running and waiting jobs
//   - GET <route>/logs returns the latest log lines
//   - POST <route>/{id}/stop stops a job
func (j *Josuke) HandleJobs(logs *LogBuffer) {
	if j.JobsAPI == nil {
		return
	}
	if j.JobsAPI.Token == "" {
		log.Fatal("[ERR ] jobs_api requires a token")
	}
	route := j.registerJobsAPI(http.DefaultServeMux, logs)
	if j.LogEnabled(InfoLevel) {
		log.Printf("[INFO] jobs API available on %s\n", route)
	}
}

func (j *Josuke) registerJobsAPI(mux *http.ServeMux, logs *LogBuffer) string {
	route := cmp.Or(strings.TrimSuffix(j.JobsAPI.Route, "/"), "/jobs")
	mux.HandleFunc("GET "+route, j.authorized(j.listJobs))
	mux.HandleFunc("GET "+route+"/logs", j.authorized(listLogs(logs)))
	mux.HandleFunc("POST "+route+"/{id}/stop", j.authorized(j.stopJob))
	return route
}

// authorized rejects requests without the API token, anybody reaching josuke could stop builds otherwise.
func (j *Josuke) authorized(handler http.HandlerFunc) http.HandlerFunc {
	expected := []byte("Bearer " + j.JobsAPI.Token)
	return func(rw http.ResponseWriter, req *http.Request) {
		if subtle.ConstantTimeCompare([]byte(req.Header.Get("Authorization")), expected) != 1 {
			http.Error(rw, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler(rw, req)
	}
}

func (j *Josuke) listJobs(rw http.ResponseWriter, req *http.Request) {
	writeJSON(rw, http.StatusOK, j.queue.list())
}

// listLogs returns the lines logged after the "after" line number of the "boot_id" boot,
// or every line kept if they are missing.
func listLogs(logs *LogBuffer) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		after, err := strconv.ParseUint(cmp.Or(query.Get("after"), "0"), 10, 64)
		if err != nil {
			http.Error(rw, "invalid after line number", http.StatusBadRequest)
			return
		}
		writeJSON(rw, http.StatusOK, logs.Since(query.Get("boot_id"), after))
	}
}

func (j *Josuke) stopJob(rw http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseUint(req.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(rw, "invalid job id", http.StatusBadRequest)
		return
	}
	log.Printf("job#%d [WARN] stop requested by %s\n", id, req.RemoteAddr)

	status, ok := j.queue.stop(id)
	switch {
	case !ok:
		http.Error(rw, "no such job, it may be finished already", http.StatusNotFound)
	case status == "running":
		// The command gets some time to exit, see waitDelay.
		writeJSON(rw, http.StatusAccepted, map[string]any{"id": id, "status": "stopping"})
	default:
		writeJSON(rw, http.StatusOK, map[string]any{"id": id, "status": "removed"})
	}
}

func writeJSON(rw http.ResponseWriter, status int, v any) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	if err := json.NewEncoder(rw).Encode(v); err != nil {
		log.Printf("[ERR ] could not write the response: %s\n", err)
	}
}
