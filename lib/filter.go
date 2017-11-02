package lib

import (
	"encoding/json"
	"strings"
	"github.com/tolleiv/jsonpath"
	"fmt"
)

type Filter struct {
	Name      string         `json:"name"`
	Condition string         `json:"condition"`
	Actions   []string       `json:"actions"`
	Values    []FilterValues `json:"values"`
}

type FilterValues struct {
	Name     string `json:"name"`
	JsonPath string `json:"jsonPath"`
}

func (v *FilterValues) Extract(in string) string {
	var json_data interface{}
	json.Unmarshal([]byte(in), &json_data)
	res, err := jsonpath.JsonPathLookup(json_data, v.JsonPath)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", res)
}

func (f *Filter) Match(in string) bool {
	r := leftCompareDeeper([]byte(f.Condition), []byte(in))
	return r
}

func leftCompareDeeper(a, b []byte) bool {
	var aa, bb map[string]*json.RawMessage
	err := json.Unmarshal(a, &aa)
	if err != nil {
		return false
	}
	err = json.Unmarshal(b, &bb)
	if err != nil {
		return false
	}
	return leftCompareObject(aa, bb)
}

func leftCompareBroader(a, b []byte) bool {
	var aa, bb []*json.RawMessage
	err := json.Unmarshal(a, &aa)
	if err != nil {
		return false
	}
	err = json.Unmarshal(b, &bb)
	if err != nil {
		return false
	}
	res := true
	for _, va := range aa {
		ra := false
		for _, vb := range bb {
			ra = ra || strings.Compare(string(*va), string(*vb)) == 0
			ra = ra || leftCompareDeeper(*va, *vb)
		}
		res = res && ra
	}
	return res
}

func leftCompareObject(a, b map[string]*json.RawMessage) bool {
	res := true
	for k, v := range a {
		if b[k] == nil && v != nil {
			res = false
			break
		}
		if strings.Compare(string(*v), string(*b[k])) == 0 {
			continue
		}
		res = res && (leftCompareDeeper(*v, *b[k]) || leftCompareBroader(*v, *b[k]))
	}
	return res
}
