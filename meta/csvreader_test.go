package meta

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestNewCSVReaderSupportsCommonLineEndings(t *testing.T) {
	want := [][]string{
		{"name", "description"},
		{"tent", "Family, 17.5m2"},
		{"bucket", "20 litre"},
	}
	tests := map[string]string{
		"LF":   "name,description\ntent,\"Family, 17.5m2\"\nbucket,20 litre\n",
		"CRLF": "name,description\r\ntent,\"Family, 17.5m2\"\r\nbucket,20 litre\r\n",
		"CR":   "name,description\rtent,\"Family, 17.5m2\"\rbucket,20 litre\r",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			records, err := NewCSVReader(strings.NewReader(input)).ReadAll()
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if !reflect.DeepEqual(records, want) {
				t.Errorf("records = %#v, want %#v", records, want)
			}
		})
	}
}

func TestNewCSVReaderKeepsMalformedQuotesStrict(t *testing.T) {
	_, err := NewCSVReader(strings.NewReader("name,description\nitem,bare\"quote\n")).ReadAll()
	if err == nil || err == io.EOF {
		t.Fatalf("expected malformed-quote error, got %v", err)
	}
}
