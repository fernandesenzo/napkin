package napkin

import (
	"errors"
	"regexp"
	"time"
)

type Napkin struct {
	Code string
	Text string
}

const CodeLength = 6
const MaxContentLength = 200
const DefaultTTL = time.Hour * 24

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateCode(code string) error {
	if len(code) != CodeLength {
		return ErrInvalidCode
	}
	if !codePattern.MatchString(code) {
		return ErrInvalidCode
	}
	return nil
}
func ValidateContent(content string) error {
	if len(content) > MaxContentLength {
		return ErrContentTooLong
	}
	return nil
}

var ErrInvalidCode = errors.New("Invalid code")
var ErrNapkinDoesNotExist = errors.New("no napkin with such code")
var ErrContentTooLong = errors.New("napkin exceeded maximum length")
