# Convert Factory Function into `github.com/samber/do/v2.Provider`

📦 A utility for converting factory function into `do.Provider` for integration with the [`samber/do/v2`](https://github.com/samber/do/tree/v2-%F0%9F%9A%80) DI.

## 📌 Installation

```sh
go get github.com/d-enk/dofactory
```

## 🚀 Usage

### 1. Define a Factory Function

A factory function is a function that creates instances of an object:

```go
package main

import (
    "fmt"
    "github.com/samber/do/v2"
    "github.com/d-enk/dofactory"
)

type Service struct {
    Name string
}

func NewService(name string) *Service {
    return &Service{Name: name}
}
```

### 2. Convert the Factory into a Provider

With `dofactory.ToProvider`, you can easily register a factory as a provider in the DI container:

```go
func main() {
    injector := do.New()

    // Provide service name
    do.ProvideValue(injector, "MyService")

    // Convert factory to provider
    factoryProvider := dofactory.ToProvider[*Service](NewService)

    // Register provider
    do.Provide[*Service](injector, factoryProvider)

    // Retrieve the instance from the container
    service := do.MustInvoke[*Service](injector)

    fmt.Println(service.Name) // Output: MyService
}
```

More [examples](./example_test.go)

## 🎯 How It Works

The `dofactory.ToProvider` function takes a factory function (e.g., `func() *Service`) and converts it into a `do.Provider[T]`.
The provider automatically resolves dependencies via the `do.Injector` DI container.

## 🔧 Supported Factory Function Signatures

```go
func(A, B, ...) (T, error)
func(A, B, ...) T
```

❗Variadic Functions not supported

## 📜 License

This project is licensed under the **Apache License 2.0**. See the [LICENSE](./LICENSE) file for details.
