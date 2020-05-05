package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"

	shipcapsv1beta1 "github.com/redradrat/shipcaps/api/v1beta1"
)

type CapValue struct {
	// Value holds the actual value.
	Value interface{} `json:"value"`

	// TransformationIdentifier identifies the replacement placeholder.
	TargetIdentifier shipcapsv1beta1.TargetIdentifier `json:"targetId"`
}

func CapValueMap(vals []CapValue) map[string]interface{} {
	outmap := make(map[string]interface{})
	for _, entry := range vals {
		outmap[string(entry.TargetIdentifier)] = entry.Value
	}
	return outmap
}

type AppValue struct {
	// Key refers to the Cap input key that we want to set
	Key string `json:"key"`

	// Value holds the actual value.
	Value interface{} `json:"value"`
}

func AppValuesMap(values []AppValue) (map[string]interface{}, error) {
	outmap := make(map[string]interface{})
	for _, v := range values {
		outmap[v.Key] = v.Value
	}
	return outmap, nil
}

func UnmarshalAppValues(in json.RawMessage) ([]AppValue, error) {
	avu := AppValueUnmarshaler{}
	if in != nil {
		if err := json.Unmarshal(in, &avu); err != nil {
			return nil, err
		}
	}
	return avu.Data, nil
}

// MergedCapValues returns the definitive list of CapValues resulting from matching CapInputs and AppValues and the
// provided CapValues.
func MergedCapValues(cap shipcapsv1beta1.Cap, app shipcapsv1beta1.App, log logr.Logger) ([]CapValue, error) {
	var outlist []CapValue

	avs, err := UnmarshalAppValues(app.Spec.Values)
	if err != nil {
		return nil, err
	}

	mergedAppValues, err := MergeAppValues(cap.Spec.Inputs, avs)
	if err != nil {
		return nil, err
	}
	outlist = append(outlist, mergedAppValues...)

	unmarshaler := CapValueUnmarshaler{}
	if cap.Spec.Values != nil {
		log.V(1).Info(fmt.Sprintf("unmarshaling string %s", string(cap.Spec.Values)))
		if err := json.Unmarshal(cap.Spec.Values, &unmarshaler); err != nil {
			return nil, err
		}
	}

	outlist = append(outlist, unmarshaler.Data...)
	return outlist, nil
}

// MapValues returns a list of the matched values, while matching the required inputs with the provided AppValues
func MergeAppValues(ins shipcapsv1beta1.CapInputs, appValues []AppValue) ([]CapValue, error) {

	avMap, err := AppValuesMap(appValues)
	if err != nil {
		return nil, err
	}
	var outlist []CapValue
	for _, in := range ins {
		var err bool
		data, found := avMap[in.Key]
		if !found {
			if !in.Optional {
				return nil, fmt.Errorf("required key '%s' not found in App values", in.Key)
			}
			continue
		}
		switch in.Type {
		case shipcapsv1beta1.StringInputType:
			if _, ok := data.(string); !ok {
				err = true
			}
		case shipcapsv1beta1.StringListInputType:
			if _, ok := data.([]string); !ok {
				err = true
			}
		case shipcapsv1beta1.IntInputType:
			if _, ok := data.(int); !ok {
				err = true
			}
		case shipcapsv1beta1.FloatInputType:
			if _, ok := data.(float32); !ok {
				err = true
			}
		case shipcapsv1beta1.PasswordInputType:
			if _, ok := data.(string); !ok {
				err = true
			}
		}
		if err {
			return nil, fmt.Errorf("required input '%s' is not of type '%s'", in.Key, in.Type)
		}
		outlist = append(outlist, CapValue{TargetIdentifier: in.TargetIdentifier, Value: data})
	}

	return outlist, nil
}

type CapValueUnmarshaler struct {
	Data []CapValue
}

type AppValueUnmarshaler struct {
	Data []AppValue
}

func (cu *CapValueUnmarshaler) UnmarshalJSON(b []byte) error {
	umList := []CapValue{}
	err := json.Unmarshal(b, &umList)
	if err != nil {
		return err
	}
	cu.Data = umList
	return nil
}

func (cu *AppValueUnmarshaler) UnmarshalJSON(b []byte) error {
	umList := []AppValue{}
	err := json.Unmarshal(b, &umList)
	if err != nil {
		return err
	}
	cu.Data = umList
	return nil
}
