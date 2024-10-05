# lesiw.io/ops: A framework for project operations.

## Example

Ensure the [Go toolchain][go] is installed.

Add a file to `.ops/main.go`, for example:

```go
package main

import "lesiw.io/ops"

type Ops struct{}

func main()        { ops.Handle(Ops{}) }
func (Ops) Hello() { println("Hello world!") }
```

Then use `op` to run it.

```shell
go install lesiw.io/op@latest # Install op.
op -l                         # => hello
op hello                      # => Hello world!
```

You can also play with a basic example on the [Go playground][play].

## Error handling

Op functions can be of type `func()` or `func() error`. If a `func()` op panics,
that panic will be printed as if it were an error. If a `func() error` op
panics, it is treated as a true panic and will print a stacktrace.

## Post handler functions

`ops.Defer(func())` will run `func()` at the end of the program regardless of
whether an op completes successfully.

`ops.After(func())` will run `func()` after all ops have completed successfully.

## What's the difference between this and Magefiles?

By reflecting on a type, `ops.Handle(any)` allows the use of embedding to
compose ops from multiple modules together. `ops.Handle(any)` follows the same
rules as [selectors][selectors] in the Go spec with one notable exception:
multiple methods with the same name at the same depth will be run sequentially
in the same order as their embedded types.

[go]: https://go.dev/doc/install
[play]: https://go.dev/play/p/YcUCt5RLoPR
[selectors]: https://go.dev/ref/spec#Selectors
