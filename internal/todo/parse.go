package todo

// Parse parses a single todo.txt line, including GoDo's id: and mod: tags.
//
// Parse never sets Archived or Version, since neither is encoded in the line.
// It returns an error if the line is blank or a GoDo tag is malformed (for
// example, a mod: value that isn't an RFC 3339 timestamp).
//
// Formatting the result with String must reproduce the input line, apart from
// moving id: and mod: to the end and normalizing whitespace.
func Parse(line string) (Todo, error) {
	panic("todo.Parse: not implemented")
}

// Line formats t as a todo.txt line without the id: and mod: tags. This is the
// form sent over the wire in sync requests.
func (t Todo) Line() string {
	panic("todo.Line: not implemented")
}

// String formats t as a full todo.txt line, with id: and mod: appended when
// they are set. This is the form written to todo.txt and done.txt.
func (t Todo) String() string {
	panic("todo.String: not implemented")
}
