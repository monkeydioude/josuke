package josuke

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBoot = time.Date(2026, 10, 10, 13, 0, 0, 0, time.UTC)

func writeLines(t *testing.T, b *LogBuffer, texts ...string) {
	t.Helper()
	for _, text := range texts {
		n, err := b.Write([]byte(text + "\n"))
		require.NoError(t, err)
		require.Equal(t, len(text)+1, n)
	}
}

func Test_LogBuffer_Returns_The_Lines_After_A_Line_Number(t *testing.T) {
	b := NewLogBuffer(10, testBoot)
	writeLines(t, b, "one", "two", "three")
	bootID := b.Since("", 0).BootID

	page := b.Since(bootID, 1)

	assert.Equal(t, uint64(3), page.Last)
	assert.Equal(t, []LogLine{{2, "two"}, {3, "three"}}, page.Lines)
	assert.Empty(t, b.Since(bootID, 3).Lines)
}

func Test_LogBuffer_Keeps_The_Last_Lines_Only(t *testing.T) {
	b := NewLogBuffer(2, testBoot)
	writeLines(t, b, "one", "two", "three")

	page := b.Since("", 0)

	assert.Equal(t, uint64(3), page.Last)
	assert.Equal(t, []LogLine{{2, "two"}, {3, "three"}}, page.Lines)
}

func Test_LogBuffer_Returns_Every_Line_To_A_Reader_Of_Another_Boot(t *testing.T) {
	previous := NewLogBuffer(10, testBoot.Add(-time.Hour))
	b := NewLogBuffer(10, testBoot)
	writeLines(t, b, "one", "two")
	require.NotEqual(t, previous.Since("", 0).BootID, b.Since("", 0).BootID)

	page := b.Since(previous.Since("", 0).BootID, 1)

	assert.Equal(t, []LogLine{{1, "one"}, {2, "two"}}, page.Lines)
}

func Test_LogBuffer_Keeps_A_Multiline_Entry_Whole(t *testing.T) {
	b := NewLogBuffer(10, testBoot)

	log.New(b, "job#1 ", log.Lmsgprefix).Printf("[ERR ] panic: boom\n%s", "stack line 1\nstack line 2\n")

	assert.Equal(t, []LogLine{{1, "job#1 [ERR ] panic: boom\nstack line 1\nstack line 2"}}, b.Since("", 0).Lines)
}

func Test_LogBuffer_Does_Not_Share_Its_Lines(t *testing.T) {
	b := NewLogBuffer(3, testBoot)
	writeLines(t, b, "one")
	page := b.Since("", 0)

	for i := range 5 {
		writeLines(t, b, fmt.Sprint(i))
	}

	assert.Equal(t, []LogLine{{1, "one"}}, page.Lines)
}
