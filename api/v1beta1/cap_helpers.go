package v1beta1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
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
	tpl, err := template.New("source").Funcs(sprig.FuncMap()).Parse(string(source.InLine))
	if err != nil {
		return uList, err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, values.Map()); err != nil {
		return uList, err
	}

	renderedJson := buf.Bytes()
	newparsed := unstructured.UnstructuredList{}
	if err := json.Unmarshal(renderedJson, &newparsed.Items); err != nil {
		return uList, err
	}

	return newparsed, nil
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
