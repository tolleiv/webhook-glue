package lib

type Action struct {
	Name   string        `json:"name"`
	Params []ActionParam
	Script []string      `json:"script"`
}
type ActionParam struct {
	Name  string
	Value string
}
