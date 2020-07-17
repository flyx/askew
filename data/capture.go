package data

// EventHandling describes how the event's default action should be handled.
type EventHandling int

const (
	// PreventDefault prevents the event's default action
	PreventDefault EventHandling = iota
	// DontPreventDefault doesn't prevent the event's default action
	DontPreventDefault
	// AskPreventDefault prevents the event's default action if the handler
	// returns `true`. This can only be used with handlers returning bool.
	AskPreventDefault
	// AutoPreventDefault is only valid on UnboundEventMapping and will be
	// resolved to DontPreventDefault if the handler returns void and
	// AskPreventDefault if the handler returns bool.
	AutoPreventDefault
)

// BoundParam is a capture callback parameter that is bound to a value.
type BoundParam struct {
	Param string
	Value BoundValue
}

// EventMapping describes the mapping of a DOM node's event to a declared handler.
type EventMapping struct {
	Event          string
	Handler        string
	FromController bool
	ParamMappings  []BoundParam
	Handling       EventHandling
}

// UnboundEventMapping describes an event mapping for which the parameter names
// have not been bound yet.
type UnboundEventMapping struct {
	Event         string
	Handler       string
	ParamMappings map[string]BoundValue
	Handling      EventHandling
}
