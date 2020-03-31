package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) setValidatorTombstoned(ctx sdk.Context, valAddr sdk.ValAddress,Tombstoned bool){
	//hook just a callback, so just ignore return
	validator := k.sk.Validator(ctx, valAddr)
	if validator == nil {
		return
	}

	consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())

	info, found := k.GetValidatorSigningInfo(ctx, consAddr)
	if !found {
		return
	}

	//update Tombstoned
	info.Tombstoned = Tombstoned
}