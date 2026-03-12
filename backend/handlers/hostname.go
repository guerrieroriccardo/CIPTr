package handlers

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildHostname formats a hostname from a label, number, and format configuration.
func BuildHostname(label string, num int, f HostnameFormat) string {
	numStr := fmt.Sprintf("%0*d", f.NumDigits, num)
	switch f.PrefixPosition {
	case "after":
		return numStr + label
	case "none":
		return numStr
	default: // "before"
		return label + numStr
	}
}

// ParseHostnameNumber extracts the numeric part from an existing hostname.
// Returns the number and true if successful, or 0 and false otherwise.
func ParseHostnameNumber(hostname, label string, f HostnameFormat) (int, bool) {
	var numPart string
	switch f.PrefixPosition {
	case "before":
		if !strings.HasPrefix(hostname, label) {
			return 0, false
		}
		numPart = hostname[len(label):]
	case "after":
		if !strings.HasSuffix(hostname, label) {
			return 0, false
		}
		numPart = hostname[:len(hostname)-len(label)]
	case "none":
		numPart = hostname
	default:
		return 0, false
	}

	num, err := strconv.Atoi(numPart)
	if err != nil {
		return 0, false
	}
	return num, true
}

// MaxHostnameNumber returns the maximum number for the given digit count.
func MaxHostnameNumber(digits int) int {
	n := 1
	for i := 0; i < digits; i++ {
		n *= 10
	}
	return n - 1
}

// SanitizeLabel removes spaces from a label for use in hostnames.
func SanitizeLabel(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// HostnameLikePattern returns the SQL LIKE pattern for finding existing hostnames.
func HostnameLikePattern(label string, f HostnameFormat) string {
	switch f.PrefixPosition {
	case "before":
		return label + "%"
	case "after":
		return "%" + label
	default:
		return "%"
	}
}
