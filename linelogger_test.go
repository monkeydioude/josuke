package josuke

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LineLogger_Logs_Each_Line_Once_Complete(t *testing.T) {
	var buf bytes.Buffer
	l := newLineLogger(log.New(&buf, "", 0), "> ")

	for _, chunk := range []string{"hel", "lo\nwor", "ld\r\n", "\n", "  \n"} {
		l.Write([]byte(chunk))
	}
	assert.Equal(t, "> hello\n> world\n", buf.String())

	l.Write([]byte("no newline"))
	assert.Equal(t, "> hello\n> world\n", buf.String())
	l.Flush()
	assert.Equal(t, "> hello\n> world\n> no newline\n", buf.String())
}
