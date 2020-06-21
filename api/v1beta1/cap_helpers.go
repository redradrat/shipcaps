package v1beta1

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/redradrat/shipcaps/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func (cap *Cap) GetSpec() CapSpec {
	return cap.Spec
}

func (source *CapSource) GetUnstructuredObjects(values map[string]interface{}) (unstructured.UnstructuredList, error) {
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
func ReplacePlaceholders(in map[string]interface{}, vals map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	// Iterate through the whole map
	for key, value := range in {
		// Let's check which type our value is
		switch typedval := value.(type) {
		case string:
			if id, ok := IsFullPlaceholder(typedval); ok {
				// If our value is a placeholder string, then replace the whole value with what we get
				// from our CapValues.
				out[key] = vals[id]
			} else if placeholders, ok := IsStringPlaceholders(typedval); ok {
				// If our value is a string that contains multiple placeholders, then replace the subparts
				// with what we get from our CapValues.
				intstr := typedval
				for _, placeholder := range placeholders {
					id, _ := IsFullPlaceholder(placeholder)
					targetval, ok := vals[id].(string)
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

func (capref CapReference) IsClusterCap() bool {
	return capref.Namespace == ""
}

func (capref CapReference) NamespacedName() types.NamespacedName {
	if capref.IsClusterCap() {
		return types.NamespacedName{Name: capref.Name}
	}
	return types.NamespacedName{Name: capref.Name, Namespace: capref.Namespace}
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
