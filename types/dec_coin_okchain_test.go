package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDecCoins_ToCoinsAsc(t *testing.T) {
	d1, err := NewDecFromStr("1.23456789")
	require.NoError(t, err)
	d2, err := NewDecFromStr("0.12345678")
	require.NoError(t, err)
	dc1, dc2 := NewDecCoinFromDec("okb", d1), NewDecCoinFromDec("okb", d2)
	var dcs DecCoins
	dcs = append(dcs, dc1, dc2)
	require.Equal(t, "1.23456789okb,0.12345678okb", dcs.String())
	require.Equal(t, *NewInt(123456789).i, *dcs[0].Amount.Int)
	require.Equal(t, *NewInt(12345678).i, *dcs[1].Amount.Int)
	coins := dcs.ToCoinsAsc()
	require.Equal(t, "1.23456789okb,0.12345678okb", coins.String())
	require.Equal(t, *NewInt(123456789).i, *coins[0].Amount.i)
	require.Equal(t, *NewInt(12345678).i, *coins[1].Amount.i)
}

func TestDecCoins_StandardizeDes(t *testing.T) {
	d1, err := NewDecFromStr("12345678.9")
	require.NoError(t, err)
	d2, err := NewDecFromStr("1234567890.0")
	require.NoError(t, err)
	dc1, dc2 := NewDecCoinFromDec("okb", d1), NewDecCoinFromDec("okb", d2)
	var dcs DecCoins
	dcs = append(dcs, dc1, dc2)
	require.Equal(t, *NewInt(1234567890000000).i, *dcs[0].Amount.Int)
	require.Equal(t, *NewInt(123456789000000000).i, *dcs[1].Amount.Int)
	sdcs := dcs.StandardizeDes()
	require.Equal(t, *NewInt(12345678).i, *sdcs[0].Amount.Int)
	require.Equal(t, *NewInt(1234567890).i, *sdcs[1].Amount.Int)
	require.Equal(t, "0.12345678okb,12.34567890okb", sdcs.String())
}
