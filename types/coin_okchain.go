package types

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

var (
	// Denominations can be 3 ~ 16 characters long.
	reDnmStringHp = `[a-z][a-z0-9]{0,5}(\-[a-z0-9]{3})?`
	reAmtHp       = `[[:digit:]]+`
	reDecAmtHp    = `[[:digit:]]*\.?[[:digit:]]+`
	reSpcHp       = `[[:space:]]*`
	ReDnm         = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmStringHp))
	reCoinHp      = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmtHp, reSpcHp, reDnmStringHp))
	reDecCoinHp   = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmtHp, reSpcHp, reDnmStringHp))
)
// make a copy which amount is enlarged by 10^8 times
func (coin Coin) StandardizeAsc() Coin {
	return NewCoin(coin.Denom, coin.Amount.StandardizeAsc())
}

// make a copy which amount is reduced by 10^8 times
func (coin Coin) StandardizeDes() Coin {
	return NewCoin(coin.Denom, coin.Amount.StandardizeDes())
}

// MarshalJSON marshals the coin
func (coin Coin) MarshalJSON() ([]byte, error) {
	type Alias Coin
	return json.Marshal(&struct {
		Denom  string `json:"denom"`
		Amount Dec    `json:"amount"`
	}{
		coin.Denom,
		NewDecFromIntWithPrec(coin.Amount, Precision),
	})
}

func (coin *Coin) UnmarshalJSON(data []byte) error {
	c := &struct {
		Denom  string `json:"denom"`
		Amount Dec    `json:"amount"`
	}{}
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	coin.Denom = c.Denom
	coin.Amount = NewIntFromBigInt(c.Amount.Int)
	return nil
}

// MarshalYAML marshals the coin
// showing sdk.Coins by descending power 8 with YAML decoding
func (coin Coin) MarshalYAML() (interface{}, error) {
	bytes, err := yaml.Marshal(struct {
		Denom  string
		Amount Dec
	}{
		Denom:  coin.Denom,
		Amount: NewDecFromIntWithPrec(coin.Amount, Precision),
	})

	if err != nil {
		return nil, err
	}

	return string(bytes), err
}

// ToHighPrecisionCoins is similar to ParseCoins, will parse out a list of coins separated by commas.
// but return the high-precision coins.
// If nothing is provided, it returns nil Coins.
// Returned coins are sorted.
func ToHighPrecisionCoins(coinsStr string) (coins Coins, err error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinStrs := strings.Split(coinsStr, ",")
	for _, coinStr := range coinStrs {
		coin, err := ToHighPrecisionCoin(coinStr)
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}

	// Sort coins for determinism.
	coins.Sort()

	// Validate coins before returning.
	if !coins.IsValid() {
		return nil, fmt.Errorf("parseCoins invalid: %#v", coins)
	}

	return coins, nil
}

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.

// parse a Int which is ascending by power 8 from a decimal string
// e.g: "1.23456789"  -> Int{123456789}
func ToHighPrecisionCoin(coinStr string) (coin Coin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reDecCoinHp.FindStringSubmatch(coinStr)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin expression: %s", coinStr)
	}

	denomStr, amountStr := matches[2], matches[1]

	//amount, ok := sdk.NewIntFromString(amountStr)
	amount, err := NewDecFromStr(amountStr)
	if err != nil {
		return Coin{}, fmt.Errorf("failed to parse coin amount %s: %s", amountStr, err.Error())
	}

	if err := validateDenom(denomStr); err != nil {
		return Coin{}, fmt.Errorf("invalid denom cannot contain upper case characters or spaces: %s", err)
	}

	coin = NewCoin(denomStr, NewIntFromBigInt(amount.Int))

	return coin, nil
}
