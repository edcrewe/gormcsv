package meta

import (
	"bufio"
	"encoding/csv"
	"io"
)

// NewCSVReader accepts Unix, Windows, and classic Mac record separators.
func NewCSVReader(source io.Reader) *csv.Reader {
	return csv.NewReader(&newlineReader{source: bufio.NewReader(source)})
}

type newlineReader struct {
	source *bufio.Reader
}

func (reader *newlineReader) Read(buffer []byte) (int, error) {
	for index := range buffer {
		value, err := reader.source.ReadByte()
		if err != nil {
			if index > 0 {
				return index, nil
			}
			return 0, err
		}
		if value == '\r' {
			if next, peekErr := reader.source.Peek(1); peekErr == nil && next[0] == '\n' {
				_, _ = reader.source.ReadByte()
			}
			value = '\n'
		}
		buffer[index] = value
	}
	return len(buffer), nil
}
