package mock

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type methodsMap map[string]func([]reflect.Value) []reflect.Value

func (m *methodsMap) registerMethod(name string, method func([]reflect.Value) []reflect.Value) {
	(*m)[name] = method
}

func (m *methodsMap) call(name string, args []reflect.Value) []reflect.Value {
	if method, ok := (*m)[name]; ok {
		return method(args)
	}
	return nil
}

type Ctrl[T any] struct {
	methods      *methodsMap
	callTimes    map[string]int
	returnValues map[string][]any
	lock         sync.Mutex
}

func (m *Ctrl[T]) MockReturnValues(methodName string, returnValues ...any) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.returnValues[methodName] = returnValues
}

func (m *Ctrl[T]) CallTimes(methodName string) int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.callTimes[methodName]
}

func (m *Ctrl[T]) Call(funcName string, args ...any) []any {
	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflectArgs[i] = reflect.ValueOf(arg)
	}

	reflectResults := m.methods.call(funcName, reflectArgs)
	var result []any
	for _, reflectResult := range reflectResults {
		if !reflectResult.IsValid() {
			result = append(result, nil)
		} else if reflectResult.IsZero() {
			result = append(result, reflect.Zero(reflectResult.Type()).Interface())
		} else {
			result = append(result, reflectResult.Interface())
		}
	}
	return result
}

func Gen[T any, I any](impl I) (I, *Ctrl[I]) {
	interType := reflect.TypeOf((*T)(nil)).Elem()

	methods := &methodsMap{}
	ctrl := &Ctrl[I]{
		methods:      methods,
		callTimes:    map[string]int{},
		returnValues: map[string][]any{},
	}

	mockInstance := reflect.ValueOf(impl).Elem()

	for i := 0; i < interType.NumMethod(); i++ {
		method := interType.Method(i)
		funcName := method.Name

		methods.registerMethod(funcName, func(name string) func([]reflect.Value) []reflect.Value {
			return func(args []reflect.Value) []reflect.Value {
				ctrl.lock.Lock()
				ctrl.callTimes[name]++
				ctrl.lock.Unlock()

				ctrl.lock.Lock()
				defer ctrl.lock.Unlock()
				if returnValues, ok := ctrl.returnValues[name]; ok {
					if len(returnValues) != method.Type.NumOut() {
						panic(fmt.Sprintf("number of return values for method[%s] not match", name))
					}
					reflectVals := make([]reflect.Value, len(returnValues))
					for j := range returnValues {
						reflectVals[j] = reflect.ValueOf(returnValues[j])
					}
					return reflectVals
				}
				results := make([]reflect.Value, method.Type.NumOut())
				for j := 0; j < method.Type.NumOut(); j++ {
					results[j] = reflect.Zero(method.Type.Out(j))
				}
				return results
			}
		}(funcName))
	}
	for i := 0; i < mockInstance.NumField(); i++ {
		field := mockInstance.Field(i)
		if field.Type() == reflect.TypeOf(ctrl) {
			ptrToField := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			ptrToField.Set(reflect.ValueOf(ctrl))
		}
	}

	return impl, ctrl
}
