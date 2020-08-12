package tools

func IF(condition bool, v1 interface{}, v2 interface{}) interface{} {
	if condition {
		return v1
	}
	return v2
}
