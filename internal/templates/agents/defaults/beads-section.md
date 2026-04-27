<!-- BEGIN BEADS INTEGRATION -->
## Beads & Session Completion

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Storage (overrides default bd guidance)

- Storage is a **shared Dolt SQL server hosted on Railway**, reached over Tailscale at `railway-dolt:3306`.
- **Writes are live.** Every `bd create/update/close` hits the remote server immediately. Other machines/agents see changes instantly.
- **Do NOT run `bd dolt push` or `bd dolt pull`.** The DB is already remote — there is no separate sync step.
- `BD_DOLT_AUTO_PUSH=false` and `backup.enabled: false` are intentional. Do not re-enable.
- "auto-push disabled" messaging does **not** mean local-only — it means pushing would be redundant (and was previously misconfigured against a git URL).
- Connection uses env vars in `~/.bashrc` (`BEADS_DOLT_SERVER_MODE=1` + host/port/user/password). If `bd` fails to connect, the tailnet or env vars are the issue — not local data.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Dolt-powered version control with native sync
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update <id> --claim --json
bd update bd-42 --priority 1 --json
```

**Complete work:**

```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task atomically**: `bd update <id> --claim`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Quality
- Use `--acceptance` and `--design` fields when creating issues
- Use `--validate` to check description completeness

### Lifecycle
- `bd defer <id>` / `bd supersede <id>` for issue management
- `bd stale` / `bd orphans` / `bd lint` for hygiene
- `bd human <id>` to flag for human decisions
- `bd formula list` / `bd mol pour <name>` for structured workflows

### Important Rules

- Use bd for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists.
- Always use `--json` flag for programmatic use.
- Link discovered work with `discovered-from` dependencies.
- Check `bd ready` before asking "what should I work on?"
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files.
- Do NOT create markdown TODO lists, external issue trackers, or duplicate tracking systems.

For more details, see README.md and docs/QUICKSTART.md.

### Session close protocol

Work is NOT complete until `git push` succeeds.

1. File issues for any remaining follow-up work (`bd create`).
2. Run quality gates (tests, linters, builds) if code changed.
3. Close finished issues (`bd close <id>`). bd writes are already live on Railway — **no `bd dolt push` needed**.
4. Push code:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. Verify all changes committed AND pushed before handing off. If push fails, resolve and retry — never stop with work stranded locally.

<!-- END BEADS INTEGRATION -->
