package parsing

import (
	"encoding/json"
)

type AppValues []AppValue
type CapValues []CapValue
type RawCapValues json.RawMessage
type RawAppValues json.RawMessage
type TargetIdentifier string

type CapValue struct {
	// Value holds the actual value.
	Value interface{} `json:"value"`

	// TransformationIdentifier identifies the replacement placeholder.
	TargetIdentifier TargetIdentifier `json:"targetId"`
}

func (cv *CapValues) Map() map[string]interface{} {
	outmap := make(map[string]interface{})
	for _, entry := range *cv {
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

func (av AppValues) Map() map[string]interface{} {
	outmap := make(map[string]interface{})
	for _, v := range av {
		outmap[v.Key] = v.Value
	}
	return outmap
}

func (av AppValues) Raw() (RawAppValues, error) {
	var err error
	raw := RawAppValues{}
	if len(av) != 0 {
		if raw, err = json.Marshal(av); err != nil {
			return nil, err
		}
	}
	return raw, nil
}

func (cv CapValues) Raw() (RawCapValues, error) {
	var err error
	raw := RawCapValues{}
	if len(cv) != 0 {
		if raw, err = json.Marshal(cv); err != nil {
			return nil, err
		}
	}
	return raw, nil
}

func ParseRawAppValues(in RawAppValues) (AppValues, error) {
	avu := AppValueJSON{}
	if in != nil {
		if err := json.Unmarshal(in, &avu); err != nil {
			return nil, err
		}
	}
	return avu.Data, nil
}

func ParseRawCapValues(in RawCapValues) (CapValues, error) {
	cvu := CapValueJSON{}
	if in != nil {
		if err := json.Unmarshal(in, &cvu); err != nil {
			return nil, err
		}
	}
	return cvu.Data, nil
}

type CapValueJSON struct {
	Data []CapValue
}

type AppValueJSON struct {
	Data []AppValue
}

func (cu *CapValueJSON) UnmarshalJSON(b []byte) error {
	umList := []CapValue{}
	err := json.Unmarshal(b, &umList)
	if err != nil {
		return err
	}
	cu.Data = umList
	return nil
}

func (cu *AppValueJSON) UnmarshalJSON(b []byte) error {
	umList := []AppValue{}
	err := json.Unmarshal(b, &umList)
	if err != nil {
		return err
	}
	cu.Data = umList
	return nil
}

func (cu *CapValueJSON) MarshalJSON() ([]byte, error) {
	out, err := json.Marshal(cu.Data)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (cu *AppValueJSON) MarshalJSON() ([]byte, error) {
	out, err := json.Marshal(cu.Data)
	if err != nil {
		return out, err
	}
	return out, nil
}
