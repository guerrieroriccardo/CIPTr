package resource

import (
	"fmt"
	"strconv"
	"strings"
)

// strPtr returns a pointer to s, or nil if s is empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// int64Ptr parses s as int64 and returns a pointer, or nil if s is empty.
func int64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

// intPtr parses s as int and returns a pointer, or nil if s is empty.
func intPtr(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

// boolPtr parses s as bool and returns a pointer, or nil if s is empty.
func boolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	v := s == "true" || s == "1" || s == "yes"
	return &v
}

// mustInt64 parses s as int64, returning 0 on error.
func mustInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// mustInt parses s as int, returning 0 on error.
func mustInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// derefStr returns the value of a *string, or "" if nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// derefInt64 formats a *int64, or "" if nil.
func derefInt64(p *int64) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d", *p)
}

// derefInt formats a *int, or "" if nil.
func derefInt(p *int) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d", *p)
}

// parseInt64Slice parses a comma-separated string of int64 values.
func parseInt64Slice(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// formatInt64Slice formats a slice of int64 as comma-separated string.
func formatInt64Slice(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}

// derefBool formats a *bool, or "" if nil.
func derefBool(p *bool) string {
	if p == nil {
		return ""
	}
	if *p {
		return "true"
	}
	return "false"
}
