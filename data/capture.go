package data

// EventMapping describes the mapping of a DOM node's event to a declared handler.
type EventMapping struct {
	Event         string
	Handler       string
	ParamMappings map[string]BoundValue
}
