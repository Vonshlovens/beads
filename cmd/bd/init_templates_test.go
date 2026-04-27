package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateConfigYamlRemoteServerDefaults(t *testing.T) {
	beadsDir := t.TempDir()

	if err := createConfigYaml(beadsDir, false, ""); err != nil {
		t.Fatalf("createConfigYaml: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(beadsDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}

	configYAML := string(data)
	for _, want := range []string{
		"dolt:",
		"auto-push: false",
		"backup:",
		"enabled: false",
		"Writes go directly to the remote database",
	} {
		if !strings.Contains(configYAML, want) {
			t.Errorf("config.yaml missing %q:\n%s", want, configYAML)
		}
	}
}
