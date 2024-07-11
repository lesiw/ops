package ns

import (
	"context"

	"lesiw.io/ci/stream"
)

type Namespace interface {
	StreamContext(context.Context, ...string) stream.Stream
	RunContext(context.Context, ...string) error
}
