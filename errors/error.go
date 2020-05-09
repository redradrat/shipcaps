package errors

type ShipCapsErrorCode string

type ShipCapsError struct {
	code    ShipCapsErrorCode
	message string
}

// NewShipCapsError returns a new ShipCapsError with a custom message
func NewShipCapsError(code ShipCapsErrorCode, msg string) ShipCapsError {
	return ShipCapsError{
		code:    code,
		message: msg,
	}
}

func (err ShipCapsError) Error() string {
	return err.message
}

func IsErr(err error, code ShipCapsErrorCode) bool {
	myerr, ok := err.(ShipCapsError)
	if !ok {
		return false
	}
	return myerr.code == code
}
