## Beads & Session Completion

This project uses **bd (beads)** for issue tracking. The SessionStart hook loads the full command reference automatically; run `bd prime` to reload it mid-session.

### Storage (overrides default bd guidance)

- Storage is a **shared Dolt SQL server hosted on Railway**, reached over Tailscale at `railway-dolt:3306`.
- **Writes are live.** Every `bd create/update/close` hits the remote server immediately. Other machines/agents see changes instantly.
- **Do NOT run `bd dolt push` or `bd dolt pull`.** The DB is already remote — there is no separate sync step.
- `BD_DOLT_AUTO_PUSH=false` and `backup.enabled: false` are intentional. Do not re-enable.
- "auto-push disabled" messaging does **not** mean local-only — it means pushing would be redundant (and was previously misconfigured against a git URL).
- Connection uses env vars in `~/.bashrc` (`BEADS_DOLT_SERVER_MODE=1` + host/port/user/password). If `bd` fails to connect, the tailnet or env vars are the issue — not local data.

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists.
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files.

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
