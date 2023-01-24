package util

import (
	"fmt"
	"strconv"
)

func Uint64FromStringID(id string) (uint64, error) {
	uintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid id %s, expected base-10 unsigned integer", err, id)
	}
	return uintID, nil
}
