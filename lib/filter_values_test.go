package lib

import (
	"testing"
	"strings"
)

var extractTests = []struct {
	in  string
	out string
}{
	{`{"value":"string"}`, `string`},
	{`{"notfound":"string"}`, ``},
}

func TestFilterValueExtraction(t *testing.T) {
	fv := FilterValues{"test", "$.value"}
	for _, tt := range extractTests {
		e := fv.Extract(tt.in)
		if strings.Compare(e, tt.out) != 0 {
			t.Errorf("Expected %s got '%s'", tt.out, e)
		}
	}
}
