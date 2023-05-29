package db

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/testnet/pkg/utils"
)

const FIXTURES_PATH = "../fixtures"

func TestLoadVASPs(t *testing.T) {
	vasps, err := LoadVASPs(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 4, len(vasps))
	for _, vasp := range vasps {
		identity, err := vasp.LoadIdentity()
		require.NoError(t, err, "failed to load identity for vasp fixture %s", vasp.Name)
		require.NoError(t, identity.GetLegalPerson().Validate(), "identity for vasp fixture %s must be a valid ivms101.LegalPerson", vasp.Name)
	}
}

func TestLoadWallets(t *testing.T) {
	wallets, accounts, err := LoadWallets(FIXTURES_PATH)
	require.NoError(t, err)

	require.Equal(t, 28, len(wallets))
	require.Equal(t, 28, len(accounts))
	var testnetWallets, mainnetWallets, charlieWallets int
	for _, account := range accounts {
		identity, err := account.LoadIdentity()
		require.NoError(t, err, "failed to load identity for account fixture %s", account.Name)
		require.NoError(t, identity.GetNaturalPerson().Validate(), "identity for account fixture %s must be a valid ivms101.NaturalPerson", account.Name)
		require.NotEmpty(t, account.WalletAddress, "account fixture %s must have a wallet address", account.Name)

		if account.WalletAddress[0] == 'c' {
			charlieWallets++
		} else {
			// Bitcoin testnet wallets start with m or n, mainnet wallets start with 1 or 3
			btc, err := utils.ParseBTCAddress(account.WalletAddress)
			require.NoError(t, err, "failed to parse wallet address %s for account fixture %s", account.WalletAddress, account.Name)

			if btc.IsTestnet() {
				testnetWallets++
			} else if btc.IsMainnet() {
				mainnetWallets++
			} else {
				require.Fail(t, "fixture wallet address %s is not a testnet or mainnet address", account.WalletAddress)
			}
		}
	}

	require.Equal(t, 12, testnetWallets)
	require.Equal(t, 12, mainnetWallets)
	require.Equal(t, 4, charlieWallets)
}
