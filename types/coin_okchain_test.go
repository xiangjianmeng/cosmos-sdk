package types

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestHpPreParseCoin(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"0.001okb", "0.00100000okb"},
		{"1okb", "1.00000000okb"},
		{"100000000okb", "100000000.00000000okb"},
	}
	for _, test := range testCases {
		c, err := ToHighPrecisionCoins(test.in)

		fmt.Printf("[%s][%sokb]\n", c.String(), c.AmountOf("okb"))
		require.Nil(t, err)
		require.EqualValues(t, test.out, c.String())
	}

	_, err := ToHighPrecisionCoins("0okb")
	require.NotNil(t, err)
}

func TestMarshalYAML(t *testing.T) {
	coins := NewCoins(NewCoin("okb", NewInt(102400000000)))
	out, err := yaml.Marshal(&coins)
	require.NoError(t, err)
	expectantStr := `- |
  denom: okb
  amount: "1024.00000000"
`
	require.Equal(t, expectantStr, string(out))

}
