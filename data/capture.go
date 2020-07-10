package data

// BoundParam is a capture callback parameter that is bound to a value.
type BoundParam struct {
	Param string
	Value BoundValue
}

// EventMapping describes the mapping of a DOM node's event to a declared handler.
type EventMapping struct {
	Event         string
	Handler       string
	ParamMappings []BoundParam
}

// UnboundEventMapping describes an event mapping for which the parameter names
// have not been bound yet.
type UnboundEventMapping struct {
	Event         string
	Handler       string
	ParamMappings map[string]BoundValue
}
