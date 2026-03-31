package depin

import (
	"context"
	"reflect"
)

func isPtrToStruct(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct
}

func isFieldToInject(f reflect.Value) bool {
	return f.Kind() == reflect.Interface && f.IsNil() && !isExcludedForInjection(f)
}

func isExcludedForInjection(f reflect.Value) bool {
	return f.Type() == reflect.TypeOf(new(context.Context)).Elem()
}
