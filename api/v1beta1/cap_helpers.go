package v1beta1

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/redradrat/shipcaps/errors"
	"github.com/redradrat/shipcaps/parsing"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RenderValues takes an App Object as input and uses its spec to render a complete set of CapValues
func (cap *Cap) RenderValues(app *App) (parsing.CapValues, error) {
	var outList []parsing.CapValue

	// Unmarshal given App's values.
	avs, err := parsing.ParseRawAppValues(parsing.RawAppValues(app.Spec.Values))
	if err != nil {
		return nil, err
	}

	// Go through the whole map and see if all Inputs are given, and have the right type.
	avMap := avs.Map()
	for _, in := range cap.Spec.Inputs {
		var err bool
		data, found := avMap[in.Key]
		if !found {
			if !in.Optional {
				return nil, fmt.Errorf("required key '%s' not found in App values", in.Key)
			}
			continue
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
		// Value looks good, let's put it onto our output slice.
		outList = append(outList, parsing.CapValue{TargetIdentifier: in.TargetIdentifier, Value: data})
	}

	// Unmarshal the Values from our Cap and put them onto the output slice
	cvs, err := parsing.ParseRawCapValues(parsing.RawCapValues(cap.Spec.Values))
	if err != nil {
		return nil, err
	}
	outList = append(outList, cvs...)

	return outList, nil
}

func (source *CapSource) GetUnstructuredObjects(values parsing.CapValues) (unstructured.UnstructuredList, error) {
	uList := unstructured.UnstructuredList{}
	var parseMap []map[string]interface{}

	if err := json.Unmarshal(source.InLine, &parseMap); err != nil {
		return uList, err
	}

	for _, manifest := range parseMap {
		unstruct := unstructured.Unstructured{}
		unstruct.SetUnstructuredContent(ReplacePlaceholders(manifest, values))
		uList.Items = append(uList.Items, unstruct)
	}

	return uList, nil
}

func RemarshalAndReplacePlaceholders(in json.RawMessage, vals parsing.CapValues) (map[string]interface{}, error) {
	var parseMap map[string]interface{}

	if err := json.Unmarshal(in, &parseMap); err != nil {
		return nil, err
	}

	return ReplacePlaceholders(parseMap, vals), nil
}

// ReplacePlaceholder takes a map and replaces any found placeholder string values with arbitrary values
func ReplacePlaceholders(in map[string]interface{}, vals parsing.CapValues) map[string]interface{} {
	out := make(map[string]interface{})
	// Iterate through the whole map
	for key, value := range in {
		// Let's check which type our value is
		switch value.(type) {
		case string:
			// If our value is a placeholder string, then replace the whole value with what we get
			// from our CapValues. If it's not a Placeholder, we keep the value in place.
			if id, ok := Placeholder(value.(string)); ok {
				out[key] = vals.Map()[id]
			} else {
				out[key] = value
			}
		case map[string]interface{}:
			// If our value is another map[string]interface{}, then onwards into the rabbit hole.
			out[key] = ReplacePlaceholders(value.(map[string]interface{}), vals)
		default:
			// In all other cases, we just leave the value be.
			out[key] = value
		}
	}
	return out
}

const PlaceholderRegex = `(?m){{\s+(\S*)\s*}}`

// Placeholder checks a string input whether it is a placeholder, and returns the id of it if true
func Placeholder(in string) (string, bool) {
	rgx := regexp.MustCompile(PlaceholderRegex)
	if !rgx.MatchString(in) {
		return "", false
	}

	// Now let's get our id
	ph := rgx.FindStringSubmatch(in)[1]
	// If we didn't find a string, then the input was not a placeholder
	if ph == "" {
		// No string found, return false
		return ph, false
	}

	// We got our string, so we're positive.
	return ph, true
}

// IsInLine returns true if the
func (source *CapSource) IsInLine() bool {
	return len(source.InLine) != 0
}

func (source *CapSource) IsRepo() bool {
	return source.Repo.URI != ""
}

func (source *CapSource) Check() error {
	if !source.IsInLine() && !source.IsRepo() {
		return errors.NewShipCapsError(InvalidMaterialSpecCode, "neither inline nor repo specified")
	}
	if source.IsInLine() && source.IsRepo() {
		return errors.NewShipCapsError(InvalidMaterialSpecCode, "both inline and repo specified")
	}
	return nil
}

const (
	InvalidMaterialSpecCode errors.ShipCapsErrorCode = "InvalidMaterialSpec"
)
