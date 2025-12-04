package core

type FlowContext struct {
	ChangeNum     string
	Patchset      string
	EnableContext bool
	Data          map[string]interface{}
}

func NewFlowContext() *FlowContext {
	return &FlowContext{Data: make(map[string]interface{})}
}
