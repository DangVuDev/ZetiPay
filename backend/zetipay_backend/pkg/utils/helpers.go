package utils

import (
	"strings"
)

func TrimHexPrefix(s string) string {
	return strings.TrimPrefix(s, "0x")
}