# lesiw.io/ops: A cli for project operations.

## Build

``` sh
"$(curl lesiw.io/ops | sh)"
```

## Use

Ensure the [Go runtime][go] is installed.

Install this example to `ops/main.go` in a git repository.

``` go
package main

import "lesiw.io/ops"

type Ops struct{}

func main() {
    ops.Handle(Ops{})
}

func (op Ops) Hello() {
    println("Hello world!")
}
```

Then run it.

```shell
curl lesiw.io/ops | sh  # Install op.
op -l                   # => hello
op hello                # => Hello world!
```

[go]: https://go.dev/doc/install
