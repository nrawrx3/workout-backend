package util

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

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
		log.Error().Err(err).Msg("invalid base64 encoding for given password hash")
		return false, err
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil, nil
}

func AddJsonContentHeader(w http.ResponseWriter, status int) {
	if status != 0 {
		w.WriteHeader(status)
	}
	w.Header().Add("Content-Type", "application/json")
}

func ShuffleIntRange(start, end int) []int {
	if end < start {
		panic(fmt.Errorf("end > start (%d > %d)", end, start))
	}

	count := end - start

	slice := make([]int, count)

	for i := 0; i < count; i++ {
		slice[i] = i
	}

	for end := len(slice); end > 0; end-- {
		randomIndex := rand.Intn(end)
		slice[randomIndex], slice[end-1] = slice[end-1], slice[randomIndex]
	}

	return slice
}
