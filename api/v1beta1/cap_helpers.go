package v1beta1

import (
	"encoding/json"
	"fmt"
)

// MapValues returns a list of the mapped values, and checks whether all types are assertable as specified
func (ins CapInputs) MapValues(values AppValues) (map[string]interface{}, error) {
	outmap := make(map[string]interface{})
	for _, in := range ins {
		var err bool
		var data interface{}
		if err := json.Unmarshal(values[in.Key], &data); err != nil {
			return nil, err
		}
		switch in.Type {
		case StringInputType:
			if _, ok := data.(string); !ok {
				err = true
			}
		case StringListInputType:
			if _, ok := data.([]string); !ok {
				err = true
			}
		case IntInputType:
			if _, ok := data.(int); !ok {
				err = true
			}
		case FloatInputType:
			if _, ok := data.(float32); !ok {
				err = true
			}
		case PasswordInputType:
			if _, ok := data.(string); !ok {
				err = true
			}
		}
		if err {
			return nil, fmt.Errorf("required input '%s' is not of type '%s'", in.Key, in.Type)
		}
		outmap[in.Key] = values[in.Key]
	}
	return outmap, nil
}
