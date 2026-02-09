# lesiw.io/ops: A framework for project operations.

## Example

Ensure the [Go toolchain][go] is installed.

Add a file to `.ops/main.go`, for example:

```go
package main

import (
    "context"
    "fmt"

    "lesiw.io/ops"
)

type Ops struct{}

func main() { ops.Handle(Ops{}) }

func (Ops) Hello(ctx context.Context) error {
    fmt.Println("Hello world!")
    return nil
}
```

Then use `op` to run it.

```shell
go install lesiw.io/op@latest # Install op.
op -l                         # => hello
op hello                      # => Hello world!
```

You can also play with a basic example on the [Go playground][play].

The context provided by the framework is canceled after the op completes,
making `context.AfterFunc` a natural way to register cleanup:

```go
func (o Ops) Deploy(ctx context.Context) error {
    env, err := createEnvironment()
    if err != nil {
        return err
    }
    context.AfterFunc(ctx, func() {
        env.Cleanup()
    })
    return runTests(env)
}
```

Ops can call each other directly:

```go
func (o Ops) All(ctx context.Context) error {
    if err := o.Build(ctx); err != nil {
        return err
    }
    return o.Test(ctx)
}
```

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
[play]: https://go.dev/play/p/Ff4ZL0rh3Nc
[selectors]: https://go.dev/ref/spec#Selectors
