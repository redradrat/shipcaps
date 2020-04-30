package v1beta1

import (
	"fmt"
)

func (material *CapMaterial) Check() error {
	var matType = material.Type
	switch matType {
	case ManifestsMaterialType:
		if len(material.Manifests) == 0 {
			return ErrInvalidMaterialSpec("material spec does not define any manifests")
		}
	case RepoMaterialType:
		if material.Repo.URI == "" {
			return ErrInvalidMaterialSpec("material spec does not define a repo URI")
		}
	default:
		return ErrUnknownMaterialType(fmt.Sprintf("material spec type is unknown '%s'", matType))
	}
	return nil
}

type CapErrorCode string

const (
	InvalidMaterialSpecCode CapErrorCode = "InvalidMaterialSpec"
	UnknownMaterialTypeCode CapErrorCode = "UnknownMaterialType"
)

type CapError struct {
	code    CapErrorCode
	message string
}

func (err CapError) Error() string {
	return err.message
}

func ErrInvalidMaterialSpec(msg string) CapError {
	return CapError{
		code:    InvalidMaterialSpecCode,
		message: msg,
	}
}

func ErrUnknownMaterialType(msg string) CapError {
	return CapError{
		code:    UnknownMaterialTypeCode,
		message: msg,
	}
}

func IsErr(err error, code CapErrorCode) bool {
	myerr, ok := err.(CapError)
	if !ok {
		return false
	}
	return myerr.code == code
}
