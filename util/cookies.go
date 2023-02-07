package util

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/nrawrx3/workout-backend/constants"
)

func EncodeB64ThenWriteCookie(w http.ResponseWriter, cookie http.Cookie) error {
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	if len(cookie.String()) > 4096 {
		return constants.ErrMaxSizeExceeded
	}

	http.SetCookie(w, &cookie)
	return nil
}

func ReadCookieThenDecodeB64(r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, constants.ErrInvalidValue
	}

	return value, nil
}

func EncryptThenEncodeB64ThenWriteCookie(w http.ResponseWriter, cookie http.Cookie, aesCipher AESCipher) error {
	value, err := aesCipher.Encrypt([]byte(cookie.Value))
	if err != nil {
		return fmt.Errorf("failed to encrypt cookie value: %w", err)
	}

	if len(value) > 4096 {
		return constants.ErrMaxSizeExceeded
	}

	cookie.Value = string(value)
	return EncodeB64ThenWriteCookie(w, cookie)
}

func ReadCookieDecodeB64ThenDecrypt(r *http.Request, name string, aesCipher AESCipher) (string, error) {
	encryptedBytes, err := ReadCookieThenDecodeB64(r, name)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 encoded cookie bytes: %w", err)
	}
	decryptedBytes, err := aesCipher.Decrypt(encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cookie value: %w", err)
	}
	return string(decryptedBytes), nil
}
