package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	SystemFee = "0.0125"
)

func GetSystemFee() sdk.Coin {
	return DecStringToCoins(sdk.DefaultBondDenom, SystemFee)[0]
}
func ZeroFee() sdk.Coin {
	return DecStringToCoins(sdk.DefaultBondDenom, "0")[0]
}

func GetSysFeeCoins() sdk.Coins {
	return sdk.Coins{GetSystemFee()}
}

func DecStringToCoins(denom, amount string) sdk.Coins {
	feeDec := sdk.MustNewDecFromStr(amount)
	coin := sdk.NewCoin(denom, sdk.NewIntFromBigInt(feeDec.Int))
	var coins sdk.Coins
	coins = append(coins, coin)
	return coins
}

type ValidateMsgHandler func(ctx sdk.Context, msgs []sdk.Msg) error

type IsSystemFreeHandler func(ctx sdk.Context, msgs []sdk.Msg) bool
