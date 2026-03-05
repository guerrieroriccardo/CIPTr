package auth

import (
	"os"
	"path/filepath"
	"strings"
)

func TokenPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "ciptr", "token")
}

func SaveToken(token string) error {
	p := TokenPath()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(token), 0600)
}

func LoadToken() string {
	data, err := os.ReadFile(TokenPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func ClearToken() error {
	return os.Remove(TokenPath())
}
