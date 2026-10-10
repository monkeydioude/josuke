package josuke

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogBufferSize is the number of log lines kept in memory for the jobs API.
const LogBufferSize = 1000

// LogLine is a log entry, numbered from 1 since josuke started.
type LogLine struct {
	Seq  uint64 `json:"seq"`
	Text string `json:"text"`
}

// LogPage is the answer to a request for the latest log lines.
type LogPage struct {
	// BootID identifies the josuke process: line numbers start over when it changes.
	BootID string `json:"boot_id"`
	// Last is the number of the last line written, to ask for the next lines.
	Last  uint64    `json:"last"`
	Lines []LogLine `json:"lines"`
}

// LogBuffer is an io.Writer keeping the last lines written to it, so they can be read through the jobs API.
// It receives the output of the log package, each Write is one log entry.
type LogBuffer struct {
	bootID string
	size   int

	mu    sync.Mutex
	lines []LogLine
	last  uint64
}

func NewLogBuffer(size int, bootedAt time.Time) *LogBuffer {
	return &LogBuffer{
		bootID: strconv.FormatInt(bootedAt.UnixNano(), 36),
		size:   size,
	}
}

func (b *LogBuffer) Write(p []byte) (int, error) {
	text := strings.TrimSuffix(string(p), "\n")

	b.mu.Lock()
	defer b.mu.Unlock()
	b.last++
	if len(b.lines) == b.size {
		// The dropped line is freed when append moves the lines to a new array.
		b.lines = b.lines[1:]
	}
	b.lines = append(b.lines, LogLine{Seq: b.last, Text: text})
	return len(p), nil
}

// Since returns the lines written after the line numbered after, oldest first.
// Line numbers of another boot mean nothing to this one, every line kept is returned then.
func (b *LogBuffer) Since(bootID string, after uint64) LogPage {
	b.mu.Lock()
	defer b.mu.Unlock()

	if bootID != b.bootID {
		after = 0
	}
	start, _ := slices.BinarySearchFunc(b.lines, after+1, func(l LogLine, seq uint64) int {
		return cmp.Compare(l.Seq, seq)
	})
	return LogPage{BootID: b.bootID, Last: b.last, Lines: slices.Clone(b.lines[start:])}
}
