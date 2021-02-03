package rsaoeap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"errors"
)

// RSA implements the crypto.Cipher interface using RSA public/private key algorithm
// as specified in PKCS #1. Messages are encrypted with the public key and can only be
// decrypted using the private key. RSA objects must have a public key but the private
// key is only required for decryption.
type RSA struct {
	pub  *rsa.PublicKey
	priv *rsa.PrivateKey
}

// New creates an RSA Crypto handler with the specified key pair.
func New(pub *rsa.PublicKey, priv *rsa.PrivateKey) (_ *RSA, err error) {
	if pub == nil {
		return nil, errors.New("a public key is required for RSA operations")
	}
	return &RSA{pub: pub, priv: priv}, nil
}

// Encrypt the message using the public key.
func (c *RSA) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	hash := sha512.New()
	ciphertext, err = rsa.EncryptOAEP(hash, rand.Reader, c.pub, plaintext, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// Decrypt the message using the private key.
func (c *RSA) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	if c.priv == nil {
		return nil, errors.New("private key required for decryption")
	}

	hash := sha512.New()
	plaintext, err = rsa.DecryptOAEP(hash, rand.Reader, c.priv, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// EncryptionAlgorithm returns the name of the algorithm for adding to the Transaction.
func (c *RSA) EncryptionAlgorithm() string {
	return "RSA-OAEP-512"
}
