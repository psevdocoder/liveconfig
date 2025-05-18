package liveconfig

import "strings"

func ConvertSlice[T any](items []any) ([]T, error) {
	result := make([]T, 0, len(items))
	cfg := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true, // авто-конвертация чисел/строк
		MatchName: func(mapKey, fieldName string) bool {
			// либо точно равные (без учёта регистра), либо snake_case→CamelCase
			if strings.EqualFold(mapKey, fieldName) {
				return true
			}
			// простая конверсия snake → Camel
			parts := strings.Split(mapKey, "_")
			for i, p := range parts {
				if p == "" {
					continue
				}
				parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
			return strings.Join(parts, "") == fieldName
		},
	}
	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, err
	}
	return result, dec.Decode(items)
}
