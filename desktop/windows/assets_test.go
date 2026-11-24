package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsFileURLs(t *testing.T) {
	cases := map[string]string{
		`C:\Users\Игрок\App Data\StoneLine\Game\index.html`: "file:///C:/Users/%D0%98%D0%B3%D1%80%D0%BE%D0%BA/App%20Data/StoneLine/Game/index.html",
		`C:\Users\A#B\index.html`:                           "file:///C:/Users/A%23B/index.html",
		`\\server\games\Stone Line\index.html`:              "file://server/games/Stone%20Line/index.html",
	}
	for input, expected := range cases {
		if actual := fileURL(input); actual != expected {
			t.Fatalf("%q: got %q, want %q", input, actual, expected)
		}
	}
}

func TestPayloadAndProfileSurviveUpdates(t *testing.T) {
	root := t.TempDir()
	profile := filepath.Join(root, "WebView2", "Default", "Local Storage", "profile-test")
	if err := os.MkdirAll(filepath.Dir(profile), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profile, []byte("existing-player-progress"), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := prepareGame(root)
	if err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(root, "Game", "index.html")
	stat, _ := os.Stat(index)
	second, err := prepareGame(root)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("origin changed between launches")
	}
	after, _ := os.Stat(index)
	if !stat.ModTime().Equal(after.ModTime()) {
		t.Fatal("unchanged payload was unnecessarily replaced")
	}
	if err := os.WriteFile(index, []byte("old-version"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareGame(root); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(index)
	if err != nil || !bytes.Equal(actual, gameHTML) {
		t.Fatal("embedded game was not restored")
	}
	progress, err := os.ReadFile(profile)
	if err != nil || string(progress) != "existing-player-progress" {
		t.Fatal("existing progress changed")
	}
}
