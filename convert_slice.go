package liveconfig

import (
	"fmt"
	"reflect"
	"strings"
)

// ConvertSlice преобразует слайс map[string]interface{} в слайс T
func ConvertSlice[T any](items []any) ([]T, error) {
	var result []T
	for _, item := range items {
		rawMap, ok := item.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("item is not a map[any]any: %T", item)
		}

		// Приводим ключи к string
		m := make(map[string]interface{}, len(rawMap))
		for k, v := range rawMap {
			skey, ok := k.(string)
			if !ok {
				continue // или return err, если ключ не строка
			}
			m[skey] = v
		}

		var t T
		v := reflect.ValueOf(&t).Elem()
		tType := v.Type()

		for i := 0; i < tType.NumField(); i++ {
			field := tType.Field(i)
			if !v.Field(i).CanSet() {
				continue
			}

			candidates := []string{
				field.Name,
				camelToSnake(field.Name),
			}

			for _, key := range candidates {
				if val, ok := m[key]; ok {
					fv := v.Field(i)
					valVal := reflect.ValueOf(val)
					if valVal.Type().AssignableTo(fv.Type()) {
						fv.Set(valVal)
					} else if valVal.Type().ConvertibleTo(fv.Type()) {
						fv.Set(valVal.Convert(fv.Type()))
					}
					break
				}
			}
		}
		result = append(result, t)
	}
	return result, nil
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
