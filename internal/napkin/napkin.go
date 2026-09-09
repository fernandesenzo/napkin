package napkin

import (
	"errors"
	"regexp"
)

type Napkin struct {
	Code string
	Text string
}

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateCode(code string, codeLength int) error {
	if len(code) != codeLength {
		return ErrInvalidCode
	}
	if !codePattern.MatchString(code) {
		return ErrInvalidCode
	}
	return nil
}
func ValidateContent(content string, maxContentLength int) error {
	if len(content) > maxContentLength {
		return ErrContentTooLong
	}
	return nil
}

var ErrInvalidCode = errors.New("Invalid code")
var ErrNapkinDoesNotExist = errors.New("no napkin with such code")
var ErrContentTooLong = errors.New("napkin exceeded maximum length")
