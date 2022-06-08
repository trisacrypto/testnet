package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const FIXTURES_PATH = "../fixtures"

func TestLoadVASPs(t *testing.T) {
	vasps, err := LoadVASPs(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 4, len(vasps))
}

func TestLoadWallets(t *testing.T) {
	wallets, accounts, err := LoadWallets(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 16, len(wallets))
	require.Equal(t, 16, len(accounts))
}
