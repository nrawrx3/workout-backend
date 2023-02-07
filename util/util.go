package util

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func Uint64FromStringID(id string) (uint64, error) {
	uintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid id %s, expected base-10 unsigned integer", err, id)
	}
	return uintID, nil
}

func HashPasswordBase64(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func PasswordMatchesHash(password, base64PasswordHash string) (bool, error) {
	hash, err := base64.StdEncoding.DecodeString(base64PasswordHash)
	if err != nil {
		log.Printf("Invalid base64 encoding for given password-hash: %v", err)
		return false, err
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil, nil
}