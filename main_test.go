package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

// TestExtractRepoName checks repo name extraction from various URL formats
func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/lepotatoguy/git-reproduce.git", "git-reproduce"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
		{"justrepo", "justrepo"}, // edge case
		{"", ""},                 // empty input
	}
	for _, tt := range tests {
		got := extractRepoName(tt.input)
		if got != tt.want {
			t.Errorf("extractRepoName(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

// TestChooseValue checks flag resolution logic
func TestChooseValue(t *testing.T) {
	if got := chooseValue("long", "short"); got != "long" {
		t.Errorf("chooseValue(long, short) = %q; want %q", got, "long")
	}
	if got := chooseValue("", "short"); got != "short" {
		t.Errorf("chooseValue(\"\", short) = %q; want %q", got, "short")
	}
	if got := chooseValue("", ""); got != "" {
		t.Errorf("chooseValue(\"\", \"\") = %q; want empty string", got)
	}
}

// TestGetEnvOrDefault checks environment fallback logic
func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_KEY", "value")
	if got := getEnvOrDefault("TEST_KEY", "fallback"); got != "value" {
		t.Errorf("getEnvOrDefault returned %q; want %q", got, "value")
	}
	if got := getEnvOrDefault("NON_EXISTENT_KEY", "fallback"); got != "fallback" {
		t.Errorf("getEnvOrDefault returned %q; want %q", got, "fallback")
	}
}

// TestGetCommandOutputSuccess verifies simple command output
func TestGetCommandOutputSuccess(t *testing.T) {
	out, err := getCommandOutput("echo", "hello")
	if err != nil {
		t.Fatalf("getCommandOutput failed: %v", err)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Errorf("getCommandOutput returned %q; want %q", out, "hello")
	}
}

// TestGetCommandOutputFailure checks error path on invalid command
func TestGetCommandOutputFailure(t *testing.T) {
	_, err := getCommandOutput("nonexistentcommand123", "")
	if err == nil {
		t.Errorf("Expected error for invalid command, got nil")
	}
}

// TestYAMLSnapshotMarshal ensures snapshot serializes correctly
func TestYAMLSnapshotMarshal(t *testing.T) {
	snap := Snapshot{
		Repository:    "https://github.com/user/repo.git",
		Branch:        "main",
		Author:        "John Doe <john@example.com>",
		Commit:        "abc123",
		CommitMessage: "Initial commit",
		DateOfCommit:  "2024-04-01T12:00:00Z",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		GitVersion:    "git version 2.39.2",
		System: SystemInfo{
			OS:           "linux",
			Architecture: "amd64",
			Hostname:     "test-host",
		},
		Notes: "test note",
	}

	data, err := yaml.Marshal(&snap)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, "repository: https://github.com/user/repo.git") {
		t.Errorf("YAML missing repository field")
	}
	if !strings.Contains(out, "commit: abc123") {
		t.Errorf("YAML missing commit field")
	}
	if !strings.Contains(out, "notes: test note") {
		t.Errorf("YAML missing notes field")
	}
}

// TestYAMLSnapshotMarshalEmpty verifies empty snapshot serialization
func TestYAMLSnapshotMarshalEmpty(t *testing.T) {
	snap := Snapshot{}
	data, err := yaml.Marshal(&snap)
	if err != nil {
		t.Fatalf("yaml.Marshal failed for empty Snapshot: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, "repository: \"\"") {
		t.Errorf("Expected empty repository field in YAML output")
	}
}

// TestYAMLSnapshotUnmarshal verifies deserialization from YAML
func TestYAMLSnapshotUnmarshal(t *testing.T) {
	yamlData := `
repository: https://github.com/user/repo.git
branch: main
author: John Doe <john@example.com>
commit: abc123
commit_message: Initial commit
date_of_commit: 2024-04-01T12:00:00Z
timestamp: 2024-04-01T12:30:00Z
git_version: git version 2.39.2
system:
  os: linux
  architecture: amd64
  hostname: test-host
notes: test note
`
	var snap Snapshot
	err := yaml.Unmarshal([]byte(yamlData), &snap)
	if err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if snap.Repository != "https://github.com/user/repo.git" {
		t.Errorf("Unmarshal repository = %q; want %q", snap.Repository, "https://github.com/user/repo.git")
	}
	if snap.Commit != "abc123" {
		t.Errorf("Unmarshal commit = %q; want %q", snap.Commit, "abc123")
	}
	if snap.System.Hostname != "test-host" {
		t.Errorf("Unmarshal hostname = %q; want %q", snap.System.Hostname, "test-host")
	}
}

// TestSystemInfoFields checks SystemInfo population
func TestSystemInfoFields(t *testing.T) {
	info := getSystemInfo()
	if reflect.ValueOf(info).IsZero() {
		t.Errorf("SystemInfo struct is zero value")
	}
	if info.Hostname == "" {
		t.Errorf("SystemInfo.Hostname is empty")
	}
	if info.OS == "" {
		t.Errorf("SystemInfo.OS is empty")
	}
	if info.Architecture == "" {
		t.Errorf("SystemInfo.Architecture is empty")
	}
}
