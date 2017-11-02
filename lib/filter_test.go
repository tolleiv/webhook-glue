package lib

import "testing"

var matcherTests = []struct {
	exp bool
	in  string
	out string
}{
	{true, `{"a":"string"}`, `{"a":"string"}`},
	{true, `{"a":            "string"}`, `{"a":"string"}`},
	{false, `{"a":"string"}`, `{"a":2}`},
	{true, `{"a":2}`, `{"a":2}`},
	{true, `{"a":{"b": "string"}}`, `{"a":{"b": "string"}}`},
	{true, `{"a":{"b": "string"}}`, `{"a":{"b": "string"},"c": "ignored"}`},
	{false, `{"a":{"b": "string"},"c": "value"}`, `{"a":{"b": "string"}}`},
	{true, `{"a":["thing"]}`, `{"a":["thing", "thong"]}`},
	{true, `{"a":["thing"]}`, `{"a":[   "thing"]}`},
	{true, `{"a":["thing", "thong"]}`, `{"a":["thing", "thang", "thong", "thung"]}`},
	{true, `{"a":[{"c": 4}]}`, `{"a":[{"b": 3, "c": 4}]}`},
}

func TestSimpleFieldFilter(t *testing.T) {
	for _, tt := range matcherTests {
		f := Filter{Condition: tt.in}
		m := f.Match(tt.out)
		if m != tt.exp {
			t.Errorf("Expected %v got %v - input: %v %v", tt.exp, m, tt.in, tt.out)
		}
	}
}
