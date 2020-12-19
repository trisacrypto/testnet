package trisads

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"math/rand"
)

// Encrypt is a helper utility to encrypt a plain text string with the server's secret
// token, returns a cipher string which is the base64 encoded
func (s *Server) Encrypt(plaintext string) (ciphertext, signature []byte, err error) {
	// Create a 32 byte signature of the key
	hash := sha256.New()
	hash.Write([]byte(s.conf.SecretKey))
	key := hash.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce, err := GenRandom(12)
	if err != nil {
		return nil, nil, err
	}

	ciphertext = aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	if len(ciphertext) == 0 {
		return nil, nil, errors.New("could not generate aesgcm seal")
	}

	hm := hmac.New(sha256.New, key)
	hm.Write(ciphertext)
	return ciphertext, hm.Sum(nil), nil
}

// GenRandom bytes data
func GenRandom(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
