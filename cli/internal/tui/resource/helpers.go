package resource

import (
	"fmt"
	"strconv"
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
