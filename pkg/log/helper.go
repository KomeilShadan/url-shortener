package log

func MapToZapParams(extra map[ExtraKey]interface{}) []interface{} {
	params := make([]interface{}, 0)
	for key, val := range extra {
		params = append(params, string(key))
		params = append(params, val)
	}

	return params
}
