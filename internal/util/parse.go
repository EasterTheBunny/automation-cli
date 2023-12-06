package util

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrExponentParseError = fmt.Errorf("exponent parse error")
	ErrInvalidAmount      = fmt.Errorf("invalid amount")
)

func ParseExp(input string) (*big.Int, error) {
	var (
		amount *big.Int
		valid  bool
	)

	allDigits := regexp.MustCompile(`^[0-9]+$`)
	usesExp := regexp.MustCompile(`^[0-9]+e[0-9]+$`)
	strAmount := strings.TrimSpace(input)

	switch {
	case allDigits.MatchString(strAmount):
		amount, valid = new(big.Int).SetString(strAmount, 10)
		if !valid {
			return amount, fmt.Errorf("%w: invalid big int", ErrExponentParseError)
		}
	case usesExp.MatchString(strAmount):
		parts := strings.Split(strAmount, "e")

		zeros, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return amount, fmt.Errorf("%w: %s", ErrExponentParseError, err.Error())
		}

		var strVal strings.Builder

		strVal.WriteString(parts[0])

		for x := 0; x < int(zeros); x++ {
			strVal.WriteString("0")
		}

		amount, valid = new(big.Int).SetString(strVal.String(), 10)
		if !valid {
			return amount, fmt.Errorf("%w: invalid big int", ErrExponentParseError)
		}
	default:
		return amount, ErrInvalidAmount
	}

	return amount, nil
}
