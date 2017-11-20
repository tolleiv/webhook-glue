package lib

import (
	"encoding/json"
	"fmt"
	"github.com/tolleiv/jsonpath"
)

// FilterValues represent the data read from the webhook payload
type FilterValues struct {
	Name     string `json:"name"`
	JSONPath string `json:"jsonPath"`
}

// Extract reads the webhook payload and exports relevant data
func (v *FilterValues) Extract(in string) string {
	var jsonData interface{}
	json.Unmarshal([]byte(in), &jsonData)
	res, err := jsonpath.JsonPathLookup(jsonData, v.JSONPath)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", res)
}
