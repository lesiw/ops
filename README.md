# lesiw.io/ci: Command Interface

A package for ergonomic interaction with the shell, designed for automation
scripts.

## Example

``` go
package main

import (
    "lesiw.io/ci"
    "lesiw.io/ci/cmd"
)

func main() {
    defer ci.Handler()
    
    if !cmd.Check("false").Ok {
        cmd.Run("echo", "false is false")
    }
    cmd.Run("false") // Throws an error.
    cmd.Run("echo", "The script will stop before it gets here.")
}
```

## Build

``` sh
go run ./ci/main.go build
```
