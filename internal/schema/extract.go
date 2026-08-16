package schema

import "encoding/json"

func Extract(data []byte) (Schema, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	out := make(Schema)
	walk("", v, out)
	return out, nil
}

func walk(path string, v interface{}, out Schema) {
	switch val := v.(type) {

	case map[string]interface{}:
		if len(val) == 0 {
			// empty object still has a shape worth recording
			out[path] = TypeObject
			return
		}
		for k, child := range val {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}
			walk(childPath, child, out)
		}

	case []interface{}:
		arrPath := path + "[]"
		if len(val) == 0 {
			// can't infer element type from an empty array
			out[arrPath] = TypeUnknown
			return
		}
		for _, item := range val {
			walk(arrPath, item, out)
		}

	case string:
		out[path] = TypeString

	case float64: // encoding/json decodes all JSON numbers as float64
		out[path] = TypeNumber

	case bool:
		out[path] = TypeBool

	case nil:
		out[path] = TypeNull
	}
}
