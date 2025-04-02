package dofactory_test

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

func ExampleNewService() {
	injector := do.New()

	// Provide service name
	do.ProvideValue(injector, "MyService")

	// Convert factory to provider
	factoryProvider := dofactory.ToProvider[*Service](NewService)

	// Register provider
	do.Provide(injector, factoryProvider)

	// Retrieve the instance from the container
	service := do.MustInvoke[*Service](injector)

	fmt.Println(service.Name) // Output: MyService
}

func NewServiceWithError(name string, withError bool) (*Service, error) {
	if withError {
		return nil, fmt.Errorf("ERROR")
	}

	return &Service{Name: name}, nil
}

func ExampleNewServiceWithError() {
	injector := do.New()

	do.ProvideValue(injector, "MyService")

	factoryProvider := dofactory.ToProvider[*Service](NewServiceWithError)

	do.ProvideTransient(injector, factoryProvider)

	// Register flag - NO error
	do.ProvideValue(injector, false)
	fmt.Println(do.Invoke[*Service](injector))

	// Set flag - error
	do.OverrideValue(injector, true)
	fmt.Println(do.Invoke[*Service](injector))

	// output:
	// &{MyService} <nil>
	// <nil> ERROR
}
