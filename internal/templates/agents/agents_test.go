package agents

import (
	"strings"
	"testing"
)

func TestEmbeddedDefault(t *testing.T) {
	content := EmbeddedDefault()

	if content == "" {
		t.Fatal("EmbeddedDefault() returned empty string")
	}

	required := []string{
		"# Agent Instructions",
		"## Quick Reference",
		"bd prime",
		"BEGIN BEADS INTEGRATION",
		"END BEADS INTEGRATION",
		"Session close protocol",
		"git push",
	}
	for _, want := range required {
		if !strings.Contains(content, want) {
			t.Errorf("EmbeddedDefault() missing %q", want)
		}
	}
}

func TestEmbeddedBeadsSection(t *testing.T) {
	section := EmbeddedBeadsSection()

	if section == "" {
		t.Fatal("EmbeddedBeadsSection() returned empty string")
	}

	if !strings.HasPrefix(section, "<!-- BEGIN BEADS INTEGRATION -->") {
		t.Error("beads section should start with begin marker")
	}

	trimmed := strings.TrimSpace(section)
	if !strings.HasSuffix(trimmed, "<!-- END BEADS INTEGRATION -->") {
		t.Error("beads section should end with end marker")
	}

	required := []string{
		"bd create",
		"bd update",
		"bd close",
		"bd ready",
		"discovered-from",
	}
	for _, want := range required {
		if !strings.Contains(section, want) {
			t.Errorf("EmbeddedBeadsSection() missing %q", want)
		}
	}
}

func TestBeadsSectionContainsLanding(t *testing.T) {
	section := EmbeddedBeadsSection()
	if !strings.Contains(section, "Session close protocol") {
		t.Error("beads section should contain session completion content within markers")
	}
}

func TestDefaultContainsBothSections(t *testing.T) {
	content := EmbeddedDefault()

	beadsIdx := strings.Index(content, "BEGIN BEADS INTEGRATION")
	completionIdx := strings.Index(content, "Session close protocol")

	if beadsIdx == -1 {
		t.Fatal("missing beads integration section")
	}
	if completionIdx == -1 {
		t.Fatal("missing session completion section")
	}
	if beadsIdx > completionIdx {
		t.Error("beads section should come before session completion section")
	}
}
