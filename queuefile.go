package josuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

// savedQueue is the content of the queue file.
type savedQueue struct {
	LastID  uint64     `json:"last_id"`
	Running *savedJob  `json:"running,omitempty"`
	Waiting []savedJob `json:"waiting"`
}

type savedJob struct {
	ID          uint64       `json:"id"`
	Description string       `json:"description"`
	QueuedAt    time.Time    `json:"queued_at"`
	Actions     []HookAction `json:"actions"`
}

func toSavedJob(j *job) savedJob {
	return savedJob{ID: j.id, Description: j.desc, QueuedAt: j.queuedAt, Actions: j.actions}
}

// load queues again the jobs that were waiting when josuke stopped.
// The job that was running is not run again: it could be the one restarting josuke, which would loop forever.
func (q *jobQueue) load() error {
	if q.path == "" {
		return nil
	}
	data, err := os.ReadFile(q.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var saved savedQueue
	if err := json.Unmarshal(data, &saved); err != nil {
		return fmt.Errorf("could not parse %s, remove it to start with an empty queue: %w", q.path, err)
	}

	q.lastID = saved.LastID
	if r := saved.Running; r != nil {
		log.Printf("job#%d [WARN] %s was interrupted by a restart, it will not run again\n", r.ID, r.Description)
	}
	for _, s := range saved.Waiting {
		q.waiting = append(q.waiting, newJob(s.ID, s.Description, s.Actions, s.QueuedAt))
	}
	if len(q.waiting) > 0 {
		log.Printf("[INFO] restored %d waiting job(s) from %s\n", len(q.waiting), q.path)
	}
	// Forget the interrupted job, it should not be reported again at next boot.
	q.save()
	return nil
}

// save writes the queue to its file, if any. q.mu must be held.
func (q *jobQueue) save() {
	if q.path == "" {
		return
	}
	saved := savedQueue{LastID: q.lastID, Waiting: make([]savedJob, 0, len(q.waiting))}
	if q.running != nil {
		running := toSavedJob(q.running)
		saved.Running = &running
	}
	for _, j := range q.waiting {
		saved.Waiting = append(saved.Waiting, toSavedJob(j))
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err == nil {
		err = writeFileAtomic(q.path, data)
	}
	if err != nil {
		log.Printf("[ERR ] could not save the queue to %s: %s\n", q.path, err)
	}
}

// writeFileAtomic replaces the file content, so a crash never leaves it half written.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // fails once renamed, that's fine
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
