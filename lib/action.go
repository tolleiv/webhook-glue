package lib

// Action represents the scripts triggered by the matched hooks
type Action struct {
	Name   string `json:"name"`
	Params []ActionParam
	Script []string `json:"script"`
}

// ActionParam represents the values passed to the actions
type ActionParam struct {
	Name  string
	Value string
}
