package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/testnet/pkg/utils"
)

func TestParseBTCAddress(t *testing.T) {
	testCases := []struct {
		address string
		testnet bool
		err     string
	}{
		{"moJuU1GjhJzUdUGukw13a4w6CWjfFsJ92_", true, "invalid base58 character '_'"},
		{"moJuU1GjhJzUdUGukw13a4w6CWjfFsJ92qA", true, utils.ErrTooLong.Error()},
		{"1oJuU1GjhJzUdUGukw13a4w6CWjfFsJ92q", false, utils.ErrInvalidChecksum.Error()},
		{"moJuU1GjhJzUdUGukw13a4w6CWjfFsJ92q", true, ""},
		{"n1CqdbqoPZ8Y11UzQG6KyapWj4vcN3owcs", true, ""},
		{"18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh", false, ""},
	}

	for _, tc := range testCases {
		btc, err := utils.ParseBTCAddress(tc.address)
		if tc.err == "" {
			require.NoError(t, err, "expected address %s to be valid", tc.address)
			if tc.testnet {
				require.True(t, btc.IsTestnet(), "expected address %s to be testnet", tc.address)
				require.False(t, btc.IsMainnet(), "expected address %s to not be mainnet", tc.address)
			} else {
				require.False(t, btc.IsTestnet(), "expected address %s to not be testnet", tc.address)
				require.True(t, btc.IsMainnet(), "expected address %s to be mainnet", tc.address)
			}
		} else {
			require.EqualError(t, err, tc.err, "expected address %s to be invalid", tc.address)
		}
	}
}
