package dofactory

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/samber/do/v2"
	typetostring "github.com/samber/go-type-to-string"
)

type (
	testType  string
	testAlias = struct{}
)

var testTypeString = typetostring.GetType[testType]()

var (
	_int_1    = int(1)
	_int32_2  = int32(2)
	_int64_3  = int64(3)
	_string_A = string("string")
	_testType = testType("testType")
)

func TestToProvider(t *testing.T) {
	t.Parallel()

	assert := makeAssert(t)

	injector := do.New()

	assertEqual := func(got, want any) {
		t.Helper()

		assert(reflect.DeepEqual(got, want), "unexpected: got %+v, want %+v", got, want)
	}

	do.Provide(injector, ToProvider[int](func() int { return _int_1 }))
	do.Provide(injector, ToProvider[string](func() string { return _string_A }))
	do.Provide(injector, ToProvider[int32](func() (int32, error) { return _int32_2, nil }))
	do.Provide(injector, ToProvider[testType](func() (testType, error) { return _testType, nil }))

	assertEqual(do.MustInvoke[int](injector), _int_1)
	assertEqual(do.MustInvoke[string](injector), _string_A)
	assertEqual(do.MustInvoke[int32](injector), _int32_2)
	assertEqual(do.MustInvoke[testType](injector), _testType)

	type TypeError string

	do.Provide(injector, ToProvider[TypeError](func() (TypeError, error) {
		return "", fmt.Errorf("expected error")
	}))

	if _, err := do.Invoke[TypeError](injector); err.Error() != "expected error" {
		t.Errorf("unexpected error: %s", err)
	}

	type TypeNotInvoked string
	type TypeNotInvokedInput string

	do.Provide(injector, ToProvider[TypeNotInvoked](func(TypeNotInvokedInput) TypeNotInvoked { return "" }))

	if _, err := do.Invoke[TypeNotInvoked](injector); !strings.Contains(err.Error(),
		fmt.Sprintf("DI: could not find service `%s`", typetostring.GetType[TypeNotInvokedInput]())) {
		t.Errorf("unexpected error: %s", err)
	}

	type TypeNilPtr string
	type TypeNilPtrInput string
	do.ProvideTransient(injector, ToProvider[TypeNilPtr](func(in *TypeNilPtrInput) TypeNilPtr {
		if in == nil {
			return "<nil>"
		}
		return TypeNilPtr(*in)
	}))

	do.ProvideValue[*TypeNilPtrInput](injector, nil)
	assertEqual(do.MustInvoke[TypeNilPtr](injector), TypeNilPtr("<nil>"))

	var typeNilPtrInput TypeNilPtrInput = "typeNilPtrInput"
	do.OverrideValue(injector, &typeNilPtrInput)
	assertEqual(do.MustInvoke[TypeNilPtr](injector), TypeNilPtr(typeNilPtrInput))
}

func TestCast(t *testing.T) {
	t.Parallel()

	assert := makeAssert(t)

	for wantErr, factory := range map[string]Factory[any]{
		"cannot use variadic func(...int)": func(...int) {},

		"cannot use int":    0,
		"cannot use string": "",

		"cannot use func()":                          func() {},
		"cannot use func(int, string, struct {})":    func(int, string, struct{}) {},
		"cannot use func() (int, string, struct {})": func() (int, string, struct{}) { panic("") },
		"cannot use func() error":                    func() error { panic("") },
		"cannot use func() (int, error)":             func() (int, error) { panic("") },
		"cannot use func(struct {})":                 func(testAlias) {},
		"cannot use func() struct {}":                func() testAlias { panic("") },

		"cannot use func(" + testTypeString + ")": func(testType) {},
		"cannot use func() " + testTypeString:     func() testType { panic("") },
	} {
		var got string
		func() {
			defer func() { got = fmt.Sprint(recover()) }()
			cast[any](factory)
		}()

		want := wantErr + " as Factory func() (interface {}[, error])"

		assert(strings.Contains(got, want), "unexpected error: got %s, want %s", got, want)
	}

	expectedPanic := "expected panic"

	for want, factory := range map[string]reflect.Value{
		"func() int":                   cast[int](func() int { panic(expectedPanic) }),
		"func() (int, error)":          cast[int](func() (int, error) { panic(expectedPanic) }),
		"func() interface {}":          cast[any](func() any { panic(expectedPanic) }),
		"func() (interface {}, error)": cast[any](func() (any, error) { panic(expectedPanic) }),
		"func() (error, error)":        cast[error](func() (error, error) { panic(expectedPanic) }),
		"func(int) (int, error)":       cast[int](func(int) (int, error) { panic(expectedPanic) }),

		"func() struct {}":                             cast[struct{}](func() testAlias { panic(expectedPanic) }),
		"func() " + testTypeString:                     cast[testType](func() testType { panic(expectedPanic) }),
		"func(int, string, interface {}) (int, error)": cast[int](func(int, string, any) (int, error) { panic(expectedPanic) }),
	} {
		got := typetostring.GetReflectValueType(factory)
		assert(got == want, "unexpected: got %s, want %s", got, want)

		var err string

		func() {
			defer func() { err = fmt.Sprint(recover()) }()

			in := make([]reflect.Value, factory.Type().NumIn())
			for i := range in {
				in[i] = reflect.New(factory.Type().In(i)).Elem()
			}

			_ = factory.Call(in)
		}()

		assert(err == expectedPanic, "unexpected panic: got %s, want %s", err, want)
	}
}

func TestGetParametersNames(t *testing.T) {
	t.Parallel()

	assert := makeAssert(t)

	for want, fun := range map[string]any{
		"":                        func() {},
		"int,string":              func(int, string) {},
		"int,string,interface {}": func(int, string, any) any { panic("") },

		testTypeString + ",struct {}": func(testType, testAlias) any { panic("") },
	} {
		inputNames := getParametersNames(reflect.TypeOf(fun))

		got := strings.Join(inputNames, ",")
		assert(got == want, "unexpected names: got %s, want %s", got, want)
	}
}

func TestInvokeIn(t *testing.T) {
	t.Parallel()

	assert := makeAssert(t)

	injector := do.New()

	type testInvokeNamedType string

	testInvokeNamedType_A := testInvokeNamedType("A")

	do.ProvideValue(injector, _int_1)
	do.ProvideValue(injector, _int32_2)
	do.ProvideValue(injector, _int64_3)
	do.ProvideValue(injector, _string_A)
	do.ProvideValue(injector, _testType)
	do.ProvideValue(injector, testInvokeNamedType_A)

	do.ProvideValue(injector, &_int_1)

	makeInvokeIn := func(f Factory[any]) func(injector do.Injector) ([]reflect.Value, error) {
		t.Helper()

		v := reflect.ValueOf(f)
		assert(v.Kind() == reflect.Func, "unexpected Factory: %+v", f)

		return factory[any]{
			Value:           v,
			parametersNames: getParametersNames(v.Type()),
		}.invokeIn
	}

	invoke := func(f Factory[any]) (out []any) {
		t.Helper()

		gotValues, err := makeInvokeIn(f)(injector)
		assert(err == nil, fmt.Sprint(err))

		for _, v := range gotValues {
			assert(v.IsValid(), "unexpected invalid value")

			out = append(out, v.Interface())
		}

		return
	}

	check := func(eq bool, got, want []any) {
		t.Helper()
		t.Helper()

		assert(reflect.DeepEqual(got, want) == eq, "unexpected: got %+v, want %+v", got, want)
	}

	yes := func(got []any, want ...any) { check(true, got, want) }
	not := func(got []any, want ...any) { check(false, got, want) }

	yes(invoke(func(int) {}), _int_1)
	not(invoke(func(int) {}), _int64_3)
	not(invoke(func(int) {}), int(0))

	yes(invoke(func(*int) {}), &_int_1)
	not(invoke(func(*int) {}), _int_1)

	yes(invoke(func(int64) {}), _int64_3)
	not(invoke(func(int64) {}), _int_1)
	not(invoke(func(int64) {}), int(0))

	yes(invoke(func(string) {}), _string_A)
	not(invoke(func(string) {}), string(""))
	not(invoke(func(string) {}), testInvokeNamedType(""))

	yes(invoke(func(int, string) {}), _int_1, _string_A)
	yes(invoke(func(int, testType) {}), _int_1, _testType)
	yes(invoke(func(int, testInvokeNamedType) {}), _int_1, testInvokeNamedType_A)

	type Interface any
	do.ProvideValue[Interface](injector, nil)
	yes(invoke(func(Interface) {}), nil)

	do.OverrideValue(injector, Interface(0))
	yes(invoke(func(Interface) {}), 0)

	yes(invoke(func(int, int32, int64, string, testType, testInvokeNamedType) {}),
		_int_1, _int32_2, _int64_3, _string_A, _testType, testInvokeNamedType_A)

	for _, f := range []Factory[any]{
		func(complex128) {},
		func(int, complex128) {},
		func(struct{ string }) {},
	} {
		values, err := makeInvokeIn(f)(injector)

		assert(err != nil, "unexpected found: %+v", values)
		assert(strings.Contains(err.Error(), "DI: could not find service"), "unexpected error: %s", err)
	}
}

func makeAssert(t *testing.T) func(bool, string, ...any) {
	t.Helper()

	return func(cond bool, message string, v ...any) {
		t.Helper()

		if !cond {
			t.Errorf(message, v...)
		}
	}
}
