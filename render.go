package relay

import "html/template"

// Head provides a basic message status and head values
type Head struct {
	Status  int
	Content string
}

// JSON provides a basic json messages
type JSON struct {
	Head
	Data   interface{}
	Indent bool
}

// JSONP provides a basic jsonp messages
type JSONP struct {
	Head
	Callback string
	Data     interface{}
}

// HTML provides a basic html messages
type HTML struct {
	Head
	Name     string
	Template *template.Template
}

// XML provides a basic html messages
type XML struct {
	Head
	Indent bool
	Prefix []byte
}

// Text provides a basic text messages
type Text struct {
	Head
	Data string
}
