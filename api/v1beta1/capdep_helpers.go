package v1beta1

import (
	"github.com/redradrat/shipcaps/parsing"
)

// RenderValues renders the complete set of CapValues for this CapDep
func (capdep *CapDep) RenderValues() (parsing.CapValues, error) {
	// Unmarshal the Values from our Cap and put them onto the output slice
	cvs, err := parsing.ParseRawCapValues(parsing.RawCapValues(capdep.Spec.Values))
	if err != nil {
		return nil, err
	}

	return cvs, nil
}
