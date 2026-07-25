// Package idgen generates the short random IDs solar-sphere uses for users
// and devices.
package idgen

import (
	"crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
	"strconv"
	"strings"
	"time"
)

const length = 24

// New generates a 24-character pseudo-random ID. If seed is non-empty (e.g.
// an email address), its alphanumeric characters are folded in for a bit of
// human-traceability; an empty seed produces a purely random ID.
func New(seed string) string {
	var sanitized strings.Builder
	for _, ch := range strings.ToLower(seed) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			sanitized.WriteRune(ch)
		}
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomInt := int64(binary.LittleEndian.Uint64(randomBytes)) % now

	combined := sanitized.String() + strconv.FormatInt(now, 10) + strconv.FormatInt(randomInt, 10)
	if len(combined) > length {
		combined = combined[:length]
	} else if len(combined) < length {
		combined += strings.Repeat("-", length-len(combined))
	}

	runes := []rune(combined)
	for i := len(runes) - 1; i > 0; i-- {
		j := mathrand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
