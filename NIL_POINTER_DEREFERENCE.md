# Разименование nil-указателя

Разименование nil-указателя вызывает немедленную остановку программы с panic.

```go
package main

import (
	"bytes"
	"io"
)

func fn(out io.Writer) {
	if out != nil {
		out.Write([]byte("OK\n"))
	}
}

func main() {
	// case 1: nil-указатель как аргумент интерфейса
	var buf *bytes.Buffer = nil
	fn(buf) // panic

	// case 2: прямое разименование nil-указателя
	var a *int
	b := *a // panic
	println(b)
}
```

## Способ проверки

```go
// gorules/gorules.go
//go:build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// Запрещает разименование nil-указателей
func forbidNilPointerDeref(m dsl.Matcher) {
	m.Match(`*$ptr`).
		Where(m["ptr"].Text != "nil").
		Report(`possible nil pointer dereference in $ptr; add nil check before dereferencing`).
		At(m["ptr"])
}
```

```bash
$ go install github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
$ ruleguard -rules ./gorules/gorules.go .
```

```yml
# .golangci.yml
version: "2"
linters:
  enable:
    - gocritic
  settings:
    gocritic:
      enabled-checks:
        - ruleguard
      settings:
        ruleguard:
          rules: ./gorules/gorules.go
```
