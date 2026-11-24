package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const gameSHA256 = "ea9e5053fa36f198c527db8210ecbc85c51dc50dba92d69ba6733be72ec661ba"

//go:embed assets/index.html
var gameHTML []byte

//go:embed assets/THIRD-PARTY-NOTICES.txt
var thirdPartyNotices []byte

// The location is intentionally independent of executable name, release number,
// and Downloads directory, so moving/updating the EXE preserves localStorage.
func prepareGame(dataRoot string) (string, error) {
	if fmt.Sprintf("%x", sha256.Sum256(gameHTML)) != gameSHA256 {
		return "", fmt.Errorf("нарушена целостность встроенного файла игры")
	}
	gameDir := filepath.Join(dataRoot, "Game")
	if err := os.MkdirAll(gameDir, 0700); err != nil {
		return "", err
	}
	index := filepath.Join(gameDir, "index.html")
	if err := writeIfChanged(index, gameHTML); err != nil {
		return "", err
	}
	if err := writeIfChanged(filepath.Join(dataRoot, "THIRD-PARTY-NOTICES.txt"), thirdPartyNotices); err != nil {
		return "", err
	}
	return fileURL(index), nil
}

func writeIfChanged(name string, content []byte) error {
	old, err := os.ReadFile(name)
	if err == nil && bytes.Equal(old, content) {
		return nil
	}
	return os.WriteFile(name, content, 0600)
}

func fileURL(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	if strings.HasPrefix(name, "//") {
		host, path, _ := strings.Cut(strings.TrimPrefix(name, "//"), "/")
		return (&url.URL{Scheme: "file", Host: host, Path: "/" + path}).String()
	}
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return (&url.URL{Scheme: "file", Path: name}).String()
}
