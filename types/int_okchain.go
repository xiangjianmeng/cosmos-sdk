package types

import "math/big"

var (
	// 10^8
	standardDivisor = NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil))
)

// standardize Int by descending power 8
func (i Int) StandardizeDes() Int {
	return i.Quo(standardDivisor)
}

// standardize Int by ascending power 8
func (i Int) StandardizeAsc() Int {
	return i.Mul(standardDivisor)
}

// turn the Int to Dec by descending power 8
// e.g : 123456789 -> 1.23456789
func (i Int) StandardizeToDec() Dec {
	return NewDecFromIntWithPrec(i, Precision)
}
