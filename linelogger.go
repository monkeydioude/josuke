package josuke

import (
	"bytes"
	"log"
)

// maxLineLength bounds the buffer when a command writes a lot without any newline.
const maxLineLength = 64 * 1024

// lineLogger is an io.Writer logging every line written to it, as soon as it is complete.
// Blank lines are skipped.
type lineLogger struct {
	logger *log.Logger
	prefix string
	buf    []byte
}

func newLineLogger(logger *log.Logger, prefix string) *lineLogger {
	return &lineLogger{logger: logger, prefix: prefix}
}

func (l *lineLogger) Write(p []byte) (int, error) {
	l.buf = append(l.buf, p...)
	for {
		i := bytes.IndexByte(l.buf, '\n')
		if i < 0 {
			break
		}
		l.log(l.buf[:i])
		l.buf = l.buf[i+1:]
	}
	if len(l.buf) >= maxLineLength {
		l.Flush()
	}
	return len(p), nil
}

// Flush logs the last line if it was not terminated by a newline.
func (l *lineLogger) Flush() {
	l.log(l.buf)
	l.buf = l.buf[:0]
}

func (l *lineLogger) log(line []byte) {
	line = bytes.TrimRight(line, "\r")
	if len(bytes.TrimSpace(line)) == 0 {
		return
	}
	l.logger.Printf("%s%s\n", l.prefix, line)
}
