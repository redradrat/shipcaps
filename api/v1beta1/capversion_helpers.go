package v1beta1

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/redradrat/shipcaps/errors"
	"github.com/redradrat/shipcaps/parsing"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RenderValues takes an App Object as input and uses its spec to render a complete set of CapValues
func (cap *CapVersion) RenderValues(app *App) (parsing.CapValues, error) {
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
		unstructContent, err := ReplacePlaceholders(manifest, values)
		if err != nil {
			return uList, err
		}
		unstruct.SetUnstructuredContent(unstructContent)
		uList.Items = append(uList.Items, unstruct)
	}

	return uList, nil
}

// ReplacePlaceholder takes a map and replaces any found placeholder string values with arbitrary values
func ReplacePlaceholders(in map[string]interface{}, vals parsing.CapValues) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	// Iterate through the whole map
	for key, value := range in {
		// Let's check which type our value is
		switch typedval := value.(type) {
		case string:
			if id, ok := IsFullPlaceholder(typedval); ok {
				// If our value is a placeholder string, then replace the whole value with what we get
				// from our CapValues.
				out[key] = vals.Map()[id]
			} else if placeholders, ok := IsStringPlaceholders(typedval); ok {
				// If our value is a string that contains multiple placeholders, then replace the subparts
				// with what we get from our CapValues.
				intstr := typedval
				for _, placeholder := range placeholders {
					id, _ := IsFullPlaceholder(placeholder)
					targetval, ok := vals.Map()[id].(string)
					if !ok {
						return out, errors.NewShipCapsError(InvalidMaterialSpecCode, "non-string value used in in-line string replacement")
					}
					intstr = strings.ReplaceAll(intstr, placeholder, targetval)
				}
				out[key] = intstr
			} else {
				// If it's not a Placeholder, we keep the value in place.
				out[key] = value
			}
		case map[string]interface{}:
			// If our value is another map[string]interface{}, then onwards into the rabbit hole.
			intval, err := ReplacePlaceholders(typedval, vals)
			if err != nil {
				return nil, err
			}
			out[key] = intval
		default:
			// In all other cases, we just leave the value be.
			out[key] = value
		}
	}
	return out, nil
}

const FullPlaceholderRegex = `^{{\s*(\S*)\s*}}$`
const PartPlaceholderRegex = `({{\s*\S*\s*}})`

// IsFullPlaceholder checks a string input whether it is only a placeholder, and returns the id of it if true
func IsFullPlaceholder(in string) (string, bool) {
	rgx := regexp.MustCompile(FullPlaceholderRegex)
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

// IsStringPlaceholders checks a string input whether it contains placeholders, and returns the
// list of placeholders, if true
func IsStringPlaceholders(in string) ([]string, bool) {
	rgx := regexp.MustCompile(PartPlaceholderRegex)
	if !rgx.MatchString(in) {
		return nil, false
	}

	// Now let's get our placeholders
	phs := rgx.FindAllString(in, -1)
	// If we didn't find a string, then the input was not a placeholder
	if len(phs) == 0 {
		// No string found, return false
		return nil, false
	}

	// We got our strings, so we're positive.
	return phs, true
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
