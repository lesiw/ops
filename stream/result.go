package stream

type Result struct {
	Ok     bool
	Code   int
	Stream Stream
	Output string
}
