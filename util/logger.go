package util

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

func NewFileZeroLogger(filepath string) (zerolog.Logger, error) {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("failed to open file: %s: %w", filepath, err)
	}

	return zerolog.New(f), nil
}
