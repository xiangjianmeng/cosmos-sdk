package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// get commission rate from string with the limit - decimal part can't be greater than the limit numnber
func NewDecFromStrWithLimit(str string, limit int) (d Dec, err error) {
	if len(str) == 0 {
		return d, fmt.Errorf("decimal string is empty")
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return d, fmt.Errorf("decimal string is empty")
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return d, fmt.Errorf("bad decimal length")
		}
		combinedStr = combinedStr + strs[1]

	} else if len(strs) > 2 {
		return d, fmt.Errorf("too many periods to be a decimal string")
	}

	if lenDecs > limit {
		return d, fmt.Errorf(
			fmt.Sprintf("length of commission-related decimal part should not be greater than %d, len decimal %v", limit, lenDecs))
	}

	if lenDecs > Precision {
		return d, fmt.Errorf(
			fmt.Sprintf("too much precision, maximum %v, len decimal %v", Precision, lenDecs))
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr = combinedStr + zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return d, fmt.Errorf(fmt.Sprintf("bad string to integer conversion, combinedStr: %v", combinedStr))
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}
	return Dec{combined}, nil
}

