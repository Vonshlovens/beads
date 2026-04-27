## Beads & Session Completion

This project uses **bd (beads)** for issue tracking. Hook-enabled agents load the full command reference automatically; run `bd prime` to reload it mid-session.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Storage

- Storage is a shared remote Dolt SQL server, configured with `BEADS_DOLT_*` environment variables.
- Writes are live. Every `bd create/update/close` reaches the remote database immediately.
- Do NOT run `bd dolt push` or `bd dolt pull`. There is no separate beads sync step in remote-server mode.
- `BD_DOLT_AUTO_PUSH=false` and `backup.enabled: false` are intentional for this workflow.
- "auto-push disabled" messaging does not mean local-only; it means pushing would be redundant.

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items. bd writes are already live on the remote server.
4. **PUSH CODE TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
