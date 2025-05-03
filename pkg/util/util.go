package util

import "strings"

func IsBlank(value string) bool {
	s := strings.TrimSpace(value)
	return len(s) == 0
}
