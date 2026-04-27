package setup

import (
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/templates/agents"
)

func stubCodexEnvProvider(t *testing.T, env agentsEnv) {
	t.Helper()
	orig := codexEnvProvider
	codexEnvProvider = func() agentsEnv {
		return env
	}
	t.Cleanup(func() { codexEnvProvider = orig })
}

func TestInstallCodexCreatesNewFile(t *testing.T) {
	env, stdout, _ := newFactoryTestEnv(t)
	if err := installCodex(env); err != nil {
		t.Fatalf("installCodex returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Codex CLI integration installed") {
		t.Error("expected Codex install success message")
	}
	data, err := readFileBytes(env.agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(data), "profile:minimal") {
		t.Error("expected Codex setup to install the minimal profile")
	}
}

func TestCheckCodexMissingFile(t *testing.T) {
	env, stdout, _ := newFactoryTestEnv(t)
	err := checkCodex(env)
	if err == nil {
		t.Fatal("expected error for missing AGENTS.md")
	}
	if !strings.Contains(stdout.String(), "bd setup codex") {
		t.Error("expected setup guidance for codex")
	}
}

func TestInstallCodexDowngradesFullProfile(t *testing.T) {
	env, _, _ := newFactoryTestEnv(t)
	if err := os.WriteFile(env.agentsPath, []byte(agents.RenderSection(agents.ProfileFull)), 0o644); err != nil {
		t.Fatalf("write full profile: %v", err)
	}

	if err := installCodex(env); err != nil {
		t.Fatalf("installCodex returned error: %v", err)
	}

	data, err := readFileBytes(env.agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "profile:minimal") {
		t.Error("expected Codex setup to downgrade managed full profile to minimal")
	}
	if strings.Contains(content, "### Issue Types") {
		t.Error("expected Codex setup to remove verbose full profile content")
	}
}
