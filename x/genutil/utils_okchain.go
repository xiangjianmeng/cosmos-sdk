package genutil

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func highPrecisionFromInt(input sdk.Int) sdk.Int {
	return sdk.NewIntFromBigInt(new(big.Int).Mul(input.BigInt(), new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil)))
}

func toHighPrecision(coins sdk.Coins) sdk.Coins {
	if len(coins) != 1 || coins[0].Denom != sdk.DefaultBondDenom {
		panic("Invalid amount for gentx, only support 'okb'")
	}
	return sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, highPrecisionFromInt(coins[0].Amount))}
}
