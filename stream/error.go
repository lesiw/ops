package stream

type Error struct {
	err    error
	Stream Stream
	Code   int
	Log    string
}

func NewError(err error, stream Stream) error {
	if err == nil {
		return nil
	}
	return &Error{err: err, Stream: stream}
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}
