# lesiw.io/ops: A framework for project operations.

## Example

Ensure the [Go toolchain][go] is installed.

Install this file to `ops/main.go`.

```go
package main

import "lesiw.io/ops"

type Ops struct{}

func main()        { ops.Handle(Ops{}) }
func (Ops) Hello() { println("Hello world!") }
```

Then use `op` to run it.

```shell
curl lesiw.io/op | sh  # Install op.
op -l                  # => hello
op hello               # => Hello world!
```

You can also play with a basic example on the [Go playground][play].

## What's the difference between this and Magefiles?

By reflecting on a type, `ops.Handle(any)` allows the use of embedding to
compose ops from multiple modules together. `ops.Handle(any)` follows the same
rules as [selectors][selectors] in the Go spec with one notable exception:
multiple methods with the same name at the same depth will be run sequentially
in the same order as their embedded types.

[go]: https://go.dev/doc/install
[play]: https://go.dev/play/p/YcUCt5RLoPR
[selectors]: https://go.dev/ref/spec#Selectors
