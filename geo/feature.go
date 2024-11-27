package geo

type Feature map[string]interface{}

func (feature Feature) Code() string {
	props := feature["properties"].(map[string]interface{})
	if parent, ok := props["code"].(string); ok {
		return parent
	}

	return ""
}

func (feature Feature) ParentCode() string {
	props := feature["properties"].(map[string]interface{})
	if parent, ok := props["parent"].(string); ok {
		return parent
	}

	return ""
}

func (feature Feature) Name() string {
	props := feature["properties"].(map[string]interface{})
	if parent, ok := props["name"].(string); ok {
		return parent
	}

	return ""
}

func (feature Feature) Level() string {
	props := feature["properties"].(map[string]interface{})
	if parent, ok := props["level"].(string); ok {
		return parent
	}

	return ""
}
