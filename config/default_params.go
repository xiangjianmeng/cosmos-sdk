package config

import (
	"time"
)

const (

	// Update the validator set every 252 blocks by default
	DefaultBlocksPerEpoch = 252

	// Default unbonding duration, 14 days
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 2

	// Default maximum number of bonded validators
	DefaultMaxValidators uint16 = 21

	// Default maximum number of validators to vote
	DefaultMaxValsToVote = 30

	// Default minimum number of MinSelfDelegation limit by okt
	DefaultMinSelfDelegationLimit = 1

	// Default validate rate update interval by hours
	DefaultValidateRateUpdateInterval = 24
)
