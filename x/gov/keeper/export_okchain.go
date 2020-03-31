package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ConvertDecCoinsToCoins return coins by multiplying decCoins times 1e8, eg. 0.000000001okb -> 1okb
func ConvertDecCoinsToCoins(decCoins sdk.DecCoins) sdk.Coins {
	cs := make(sdk.Coins, len(decCoins))
	for i, coin := range decCoins {
		cs[i] = ConvertDecCoinToCoin(coin)
	}
	return cs
}

func ConvertDecCoinToCoin(decCoin sdk.DecCoin) sdk.Coin {
	if decCoin.Amount.LT(sdk.NewDecFromBigInt(sdk.ZeroInt().BigInt())) {
		panic(fmt.Sprintf("negative decimal coin amount: %v\n", decCoin.Amount))
	}
	if strings.ToLower(decCoin.Denom) != decCoin.Denom {
		panic(fmt.Sprintf("denom cannot contain upper case characters: %s\n", decCoin.Denom))
	}

	return sdk.Coin{
		Denom:  decCoin.Denom,
		Amount: sdk.NewIntFromBigInt(decCoin.Amount.BigInt()),
	}
}

// ConvertCoinsToDecCoins return decCoins by dividing coins by 1e8, eg. 1000000000okb -> 0.000000001okb
func ConvertCoinsToDecCoins(coins sdk.Coins) sdk.DecCoins {
	decCoins := sdk.DecCoins{}
	for _, coin := range coins {
		decCoin := sdk.NewDecCoinFromDec(coin.Denom, sdk.NewDecFromIntWithPrec(coin.Amount, sdk.Precision))
		decCoins = append(decCoins, decCoin)
	}
	decCoins = decCoins.Sort()
	return decCoins
}

func (keeper Keeper) SupplyKeeper() types.SupplyKeeper {
	return keeper.supplyKeeper
}

func (keeper Keeper) ParamSpace() types.ParamSubspace {
	return keeper.paramSpace
}

func (keeper Keeper) StoreKey() sdk.StoreKey {
	return keeper.storeKey
}

func (keeper Keeper) Cdc() types.Codec {
	return keeper.cdc
}

