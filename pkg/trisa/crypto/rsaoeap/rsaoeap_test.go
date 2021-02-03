package rsaoeap_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/testnet/pkg/trisa/crypto"
	"github.com/trisacrypto/testnet/pkg/trisa/crypto/rsaoeap"
)

func TestRSA(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	plaintext := []byte("for your eyes only -- classified")

	// Encrypt using a new cipher with just the public key
	var cipher crypto.Cipher
	cipher, err = rsaoeap.New(&priv.PublicKey, nil)

	ciphertext, err := cipher.Encrypt(plaintext)
	require.NoError(t, err)

	// Decrypt using a new cipher with both public and private key
	var decoder crypto.Cipher
	decoder, err = rsaoeap.New(&priv.PublicKey, priv)
	require.NoError(t, err)

	decoded, err := decoder.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decoded)
}
