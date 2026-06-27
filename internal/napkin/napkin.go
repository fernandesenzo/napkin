package napkin

import (
	"errors"
	"time"
)

type Napkin struct {
	Code string
	Text string
}

const CodeLength = 6
const DefaultTTL = time.Hour * 24

func ValidateCode(code string) error {
	//maybe add some more rules here?
	if len(code) != CodeLength {
		return ErrInvalidCode
	}
	return nil
}

var ErrInvalidCode = errors.New("Invalid code")
