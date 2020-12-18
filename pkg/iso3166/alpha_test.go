package iso3166_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/testnet/pkg/iso3166"
)

func TestFind(t *testing.T) {
	code, err := iso3166.Find("United States")
	require.NoError(t, err)
	require.Equal(t, "USA", code.Alpha3)

	code, err = iso3166.Find("GB")
	require.NoError(t, err)
	require.Equal(t, "United Kingdom", code.Country)

	code, err = iso3166.Find("BRA")
	require.NoError(t, err)
	require.Equal(t, "Brazil", code.Country)

	code, err = iso3166.Find("376")
	require.NoError(t, err)
	require.Equal(t, "Israel", code.Country)

	_, err = iso3166.Find("Foo")
	require.Error(t, err)
}
