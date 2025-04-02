package dofactory

import (
	"fmt"
	"reflect"

	"github.com/samber/do/v2"
	typetostring "github.com/samber/go-type-to-string"
)

type Factory[T any] any // func(...) (T[, error])

// ToProvider converts a [Factory] function into a [do.Provider].
// The returned provider can be used in a dependency injection container
// to resolve and invoke the factory function, creating instances of type T.
//
// Example:
//
//	factory := func() *Type { return &Type{} }
//	provider := dofactory.ToProvider(factory)
//	injector := do.New()
//	do.Provide(injector, provider)
//	typeInstance := do.MustInvoke[*Type](injector)
//
// Parameters:
//   - factory: The factory function that creates instances of type T.
//
// Returns:
//   - A [do.Provider] that invokes the factory function to create instances of T.
func ToProvider[T any](factory Factory[T]) do.Provider[T] {
	return newFactory(factory).provider
}

type factory[T any] struct {
	reflect.Value
	parametersNames []string
}

func newFactory[T any](_factory Factory[T]) factory[T] {
	value := cast[T](_factory)

	return factory[T]{
		Value:           value,
		parametersNames: getParametersNames(value.Type()),
	}
}

func (f factory[T]) provider(injector do.Injector) (res T, err error) {
	in, err := f.invokeIn(injector)
	if err != nil {
		return
	}

	out := f.Value.Call(in)

	res, _ = out[0].Interface().(T)

	if len(out) == 2 {
		err, _ = out[1].Interface().(error)
	}

	return
}

// invokeIn invokes a named [Factory] parameters services in the DI container as [reflect.Value].
func (f factory[_]) invokeIn(injector do.Injector) ([]reflect.Value, error) {
	values := make([]reflect.Value, len(f.parametersNames))

	for i, name := range f.parametersNames {

		v, err := do.InvokeNamed[any](injector, name)
		if err != nil {
			return nil, err
		}

		if v == nil {
			values[i] = reflect.Zero(f.Value.Type().In(i))
		} else {
			values[i] = reflect.ValueOf(v)
		}
	}

	return values, nil
}

// getParametersNames extracts the names of the input parameters of a given factory type
func getParametersNames(factoryType reflect.Type) []string {
	inputNames := make([]string, factoryType.NumIn())

	for i := range inputNames {
		inputNames[i] = typetostring.GetReflectType(factoryType.In(i))
	}

	return inputNames
}

// cast convert [Factory] to [reflect.Value] with check in/out types
func cast[T any](factory Factory[T]) reflect.Value {
	unexpected := func(v any) string {
		return fmt.Sprintf(
			"cannot use %s as Factory func() (%s[, error])",
			v, reflect.TypeOf((*T)(nil)).Elem(),
		)
	}

	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	switch {
	case factoryValue.Kind() != reflect.Func:
		panic(unexpected(factoryType))

	case factoryType.IsVariadic():
		panic(unexpected("variadic " + typetostring.GetReflectType(factoryType)))

	case factoryType.NumOut() == 2 && factoryType.Out(1) == reflect.TypeOf((*error)(nil)).Elem():
		fallthrough

	case factoryType.NumOut() == 1:
		if factoryType.Out(0) == reflect.TypeOf((*T)(nil)).Elem() {
			return factoryValue
		}
	}

	panic(unexpected(typetostring.GetReflectType(factoryType)))
}
