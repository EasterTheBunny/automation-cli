package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrExponentParseError = fmt.Errorf("exponent parse error")
	ErrInvalidAmount      = fmt.Errorf("invalid amount")
)

func ParseExp(input string) (uint64, error) {
	var (
		amount uint64
		err    error
	)

	allDigits := regexp.MustCompile(`^[0-9]+$`)
	usesExp := regexp.MustCompile(`^[0-9]+e[0-9]+$`)
	strAmount := strings.TrimSpace(input)

	switch {
	case allDigits.MatchString(strAmount):
		amount, err = strconv.ParseUint(strAmount, 10, 64)
		if err != nil {
			return amount, fmt.Errorf("%w: %s", ErrExponentParseError, err.Error())
		}
	case usesExp.MatchString(strAmount):
		parts := strings.Split(strAmount, "e")

		zeros, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return amount, fmt.Errorf("%w: %s", ErrExponentParseError, err.Error())
		}

		var strVal strings.Builder

		strVal.WriteString(parts[0])

		for x := 0; x < int(zeros); x++ {
			strVal.WriteString("0")
		}

		amount, err = strconv.ParseUint(strVal.String(), 10, 64)
		if err != nil {
			return amount, fmt.Errorf("%w: %s", ErrExponentParseError, err.Error())
		}
	default:
		return amount, ErrInvalidAmount
	}

	return amount, nil
}
