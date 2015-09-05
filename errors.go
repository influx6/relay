package relay

import "fmt"

// CustomError provides a custom error message format using a message pattern 'title: message'
type CustomError struct {
	//title of the error
	title string
	//message of the error
	message string
}

// Error returns the error message
func (c CustomError) Error() string {
	return c.String()
}

// String returns the message of the CustomError struct
func (c CustomError) String() string {
	return fmt.Sprintf("CustomError(%s): %s", c.title, c.message)
}

// NewCustomError provides a function instance generator
func NewCustomError(title, mesg string) CustomError {
	return CustomError{title, mesg}
}
