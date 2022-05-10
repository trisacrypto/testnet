package rvasp_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/testnet/pkg/rvasp"
)

const FIXTURES_PATH = "fixtures"

func TestLoadVASPs(t *testing.T) {
	vasps, err := rvasp.LoadVASPs(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 3, len(vasps))
}

func TestLoadWallets(t *testing.T) {
	wallets, accounts, err := rvasp.LoadWallets(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 12, len(wallets))
	require.Equal(t, 12, len(accounts))
}
