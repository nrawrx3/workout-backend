package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	"github.com/rs/zerolog/log"
)

// We share AEAD object created across multiple concurrent requests. The Encrypt
// and Decrypt functions are thread-safe. Only need to generate unique nonce for
// encrypting each individual message. https://github.com/golang/go/issues/41689
type AESCipher struct {
	key []byte
	// block cipher.Block
	gcm cipher.AEAD
}

func NewAESCipher(hexKey string) (*AESCipher, error) {
	if len([]byte(hexKey)) != 64 {
		return nil, ErrInvalidAESKeyLength
	}

	key, err := hex.DecodeString(hexKey)

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("failed to create cipher")
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESCipher{
		key: key,
		gcm: gcm,
	}, nil
}

var ErrInvalidAESKeyLength = errors.New("invalid AES key length")

func (a AESCipher) Encrypt(plaintextBytes []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()

	nonce := make([]byte, nonceSize)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// EXPLAIN: gcm.Seal(dst, ...) *appends* the encrypted bytes to the dst
	// array. If underlying array of dst does not have enough capacity for
	// the result, it will reallocate a new slice and copy the current
	// dst[:] to the head of that slice and then append encrypted bytes.
	// Strangely enough, Seal does *not* put the nonce as part of the bytes
	// appended. As a convention, we can append the nonce bytes and obtain
	// the final result of encryption like so: [encryptedBytes...,
	// nonceBytes...], The following allocates the dst slice to have enough
	// capacity so Seal doesn't need to reallocate.

	// EXPLAIN: Being conservative by allocating more than enough space. In
	// the cipher/gcm implementation, it looks like a `plainText + 16` bytes
	// would be the exact capacity required, but it's private code. We don't
	// usually have large messages anyway since we're using this for cookies
	encryptedLen := len(plaintextBytes) * 2

	dst := make([]byte, 0, nonceSize+encryptedLen)

	encryptedBytes := a.gcm.Seal(dst[:0], nonce, plaintextBytes, nil)

	withNonce := append(encryptedBytes, nonce...)
	return withNonce, nil
}

func (a *AESCipher) Decrypt(encryptedWithNonceBytes []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()

	encryptedBytesLen := len(encryptedWithNonceBytes) - nonceSize

	cipherText := encryptedWithNonceBytes[0:encryptedBytesLen]
	nonce := encryptedWithNonceBytes[encryptedBytesLen : encryptedBytesLen+nonceSize]
	return a.gcm.Open(nil, nonce, cipherText, nil)
}

func (aesCipher *AESCipher) MustEncryptJSON(value interface{}) io.Reader {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(value)
	if err != nil {
		log.Panic().Err(err)
	}

	encryptedBytes, err := aesCipher.Encrypt(b.Bytes())
	if err != nil {
		log.Panic().Err(err)
	}
	return bytes.NewReader(encryptedBytes)
}

func (aesCipher *AESCipher) MustDecryptJSON(source io.Reader) *json.Decoder {
	encryptedBytes, err := io.ReadAll(source)
	if err != nil {
		log.Panic().Err(err)
	}

	decryptedBytes, err := aesCipher.Decrypt(encryptedBytes)
	if err != nil {
		log.Panic().Err(err)
	}

	return json.NewDecoder(bytes.NewReader(decryptedBytes))
}
