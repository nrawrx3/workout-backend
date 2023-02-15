package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/nrawrx3/uno-backend/constants"
	"github.com/nrawrx3/uno-backend/model"
)

func EncodeB64ThenWriteCookie(w http.ResponseWriter, cookie http.Cookie, value []byte) error {
	cookie.Value = base64.URLEncoding.EncodeToString(value)

	if len(cookie.String()) > 4096 {
		return constants.ErrCodeMaxSizeExceeded
	}

	http.SetCookie(w, &cookie)
	return nil
}

func ReadCookieThenDecodeB64(r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, constants.ErrCodeNotFound
		}
		return nil, err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, constants.ErrCodeInvalidValue
	}

	return value, nil
}

func EncryptThenEncodeB64ThenWriteCookie(w http.ResponseWriter, cookie http.Cookie, aesCipher *AESCipher, value []byte) error {
	value, err := aesCipher.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt cookie value: %w", err)
	}

	if len(value) > 4096 {
		return constants.ErrCodeMaxSizeExceeded
	}

	return EncodeB64ThenWriteCookie(w, cookie, value)
}

func ReadCookieDecodeB64ThenDecrypt(r *http.Request, name string, aesCipher *AESCipher) (string, error) {
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

func ExtractSessionIDFromCookie(r *http.Request, cookieName string, aesCipher *AESCipher) (uint64, error) {
	cookieValueRaw, err := ReadCookieDecodeB64ThenDecrypt(r, cookieName, aesCipher)
	if err != nil {
		return 0, err
	}

	var cookieValue model.SessionCookieValue
	err = json.NewDecoder(strings.NewReader(cookieValueRaw)).Decode(&cookieValue)
	if err != nil {
		return 0, err
	}

	sessionId, err := Uint64FromStringID(cookieValue.SessionID)
	if err != nil {
		return 0, err
	}
	return sessionId, nil
}
