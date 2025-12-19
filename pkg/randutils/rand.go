package randutils

import (
	"math/rand/v2"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandomString(length int) string {
	var builder strings.Builder
	for range length {
		//nolint: gosec // no need to be secure
		builder.WriteByte(alphabet[rand.IntN(len(alphabet))])
	}

	return builder.String()
}
