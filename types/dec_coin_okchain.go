package types

import "math/big"

// IsAllGTE returns true iff for every denom in coins, the denom is present at
// an equal or greater amount in coinsB.
// TODO: Remove once unsigned integers are used.
func (coins DecCoins) IsAllGTE(coinsB DecCoins) bool {
	diff, _ := coins.SafeSub(coinsB)
	if len(diff) == 0 {
		return true
	}

	return !diff.IsAnyNegative()
}

func TokensFromTendermintPower(power int64) Int {
	return NewInt(power).Mul(PowerReduction)
}

// return Coins by ascending by 10^8
// eg. 1.234okb -> 123456789okb
func (coin DecCoin) ToCoinAsc() Coin {
	return NewCoin(coin.Denom, NewIntFromBigInt(coin.Amount.Int))
}

func (coins DecCoins) ToCoinsAsc() Coins {
	coinsLen := len(coins)
	cs := make(Coins, coinsLen)
	for i := 0; i < coinsLen; i++ {
		cs[i] = coins[i].ToCoinAsc()
	}
	return cs
}

// return DecCoins by descending by 10^8
// eg. 12345678.0okt -> 0.12345678okt
func (coins DecCoins) StandardizeDes() DecCoins {
	numsLen := len(coins)
	decs := make(DecCoins, numsLen)
	for i := 0; i < numsLen; i++ {
		decs[i] = DecCoin{
			Denom:  coins[i].Denom,
			Amount: Dec{new(big.Int).Quo(coins[i].Amount.Int, standardDivisor.i),},
		}
	}
	return decs
}
