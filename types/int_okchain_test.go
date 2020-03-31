package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInt_StandardizeToDec(t *testing.T) {
	i := NewInt(12345678)
	require.Equal(t, "0.12345678", i.StandardizeToDec().String())
}
