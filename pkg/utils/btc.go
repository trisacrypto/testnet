package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// All the valid characters for base58 encoding.
var tmpl = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// ParseBTCAddress decodes a base58 encoded bitcoin address and returns an error if the
// checksum is invalid.
// See https://rosettacode.org/wiki/Bitcoin/address_validation#Go
func ParseBTCAddress(address string) (btc *A25, err error) {
	btc = &A25{}
	if err = btc.Decode([]byte(address)); err != nil {
		return nil, err
	}

	if !btc.HasValidChecksum() {
		return nil, ErrInvalidChecksum
	}

	return btc, nil
}

// A25 is a type for a 25 byte (not base58 encoded) bitcoin address.
type A25 [25]byte

// Testnet addresses are assumed to start with 'm' or 'n'.
func (a *A25) IsTestnet() bool {
	return a.Version() == 110 || a.Version() == 111
}

// Mainnet addresses are assumed to start with '1' or '3'.
func (a *A25) IsMainnet() bool {
	return a.Version() == 0 || a.Version() == 5
}

// Validate an A25 checksum.
func (a *A25) HasValidChecksum() bool {
	return a.EmbeddedChecksum() == a.ComputeChecksum()
}

// Version is the first byte of the address.
func (a *A25) Version() byte {
	return a[0]
}

func (a *A25) EmbeddedChecksum() (c [4]byte) {
	copy(c[:], a[21:])
	return
}

// DoubleSHA256 computes a double sha256 hash of the first 21 bytes of the
// address. Returned is the full 32 byte sha256 hash. The bitcoin checksum will be the
// first four bytes of the slice.
func (a *A25) doubleSHA256() []byte {
	h := sha256.New()
	h.Write(a[:21])
	d := h.Sum([]byte{})
	h = sha256.New()
	h.Write(d)
	return h.Sum(d[:0])
}

// ComputeChecksum returns a four byte checksum computed from the first 21
// bytes of the address.  The embedded checksum is not updated.
func (a *A25) ComputeChecksum() (c [4]byte) {
	copy(c[:], a.doubleSHA256())
	return
}

// Decode takes a base58 encoded address and decodes it into the receiver.
// Errors are returned if the argument is not valid base58 or if the decoded
// value does not fit in the 25 byte address.  The address is not otherwise
// checked for validity.
func (a *A25) Decode(s []byte) error {
	for _, s1 := range s {
		c := bytes.IndexByte(tmpl, s1)
		if c < 0 {
			return fmt.Errorf("invalid base58 character %q", s1)
		}
		for j := 24; j >= 0; j-- {
			c += 58 * int(a[j])
			a[j] = byte(c % 256)
			c /= 256
		}
		if c > 0 {
			return ErrTooLong
		}
	}
	return nil
}
