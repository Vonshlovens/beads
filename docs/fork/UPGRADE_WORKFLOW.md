# Keeping our beads fork aligned with upstream

We maintain a small patch set on top of `gastownhall/beads` at
`Vonshlovens/beads`. Upstream ships rapidly (multiple releases per week).
This doc is the playbook for pulling new upstream work into our fork
without losing our local patches.

## Layout

- **Upstream:** `https://github.com/gastownhall/beads.git` — the source of truth.
- **Our fork:** `git@github.com:Vonshlovens/beads.git` — carries our patches on
  top of upstream `main`.
- **Local clone:** `~/dev/beads` has two remotes:
  - `origin` → our fork (push here)
  - `upstream` → `gastownhall/beads` (fetch-only)
- **Active branches on the fork:**
  - `main` — tracks `upstream/main`, plus the fork-meta docs in `docs/fork/`
    (this file and siblings). No code patches live on `main` — every code
    patch lives on its own branch so upstream rebases stay trivial.
  - `perf/bulk-wisp-partition` — the N+1 fix (see [PERFORMANCE_FIX.md](./PERFORMANCE_FIX.md)).
  - Future branches: one per logical patch, named `<topic>/<short-desc>`.
- **Installed binary:** `~/.local/bin/bd` — built locally from our fork
  (`make build` inside `~/dev/beads` then `cp ./bd ~/.local/bin/bd`).

## One-time setup on a new machine

```bash
git clone git@github.com:Vonshlovens/beads.git ~/dev/beads
cd ~/dev/beads
git remote add upstream https://github.com/gastownhall/beads.git
git fetch upstream
make build
cp ./bd ~/.local/bin/bd
```

Connection to the Railway-hosted Dolt server is configured via env vars in
your shell rc (`BEADS_DOLT_SERVER_MODE=1`, host/port/user/password) — nothing
beads-specific to set beyond that.

## Checking where we are vs. upstream

```bash
cd ~/dev/beads
git fetch upstream
git log --oneline upstream/main ^main | head       # commits upstream we don't have
git log --oneline main ^upstream/main | head       # commits of ours not upstream
```

If the second list is empty, our fork is a pure mirror — just fast-forward.

## Upgrading: the routine case (no conflicts expected)

```bash
cd ~/dev/beads
git fetch upstream

# 1. Sync our fork's main with upstream's main.
#    Our main also carries docs/fork/ on top of upstream, so the merge
#    won't fast-forward when upstream has new commits — use a rebase instead
#    so the fork-meta commits stay on top of upstream.
git checkout main
git rebase upstream/main
git push --force-with-lease origin main

# 2. Rebase each patch branch onto new upstream
for branch in perf/bulk-wisp-partition; do
    git checkout "$branch"
    git rebase main
    # Rebuild + smoke-test BEFORE force-pushing
    make build
    ./bd list --json > /dev/null   # should complete in ~5 s
    git push --force-with-lease origin "$branch"
done

# 3. Reinstall the binary on this machine
cp ~/dev/beads/bd ~/.local/bin/bd
bd --version
```

**Always use `--force-with-lease`, never bare `--force`.** `--force-with-lease`
aborts if someone else pushed to the branch in the meantime.

## Upgrading: when a rebase conflicts

Our patches touch a small surface (`internal/storage/issueops/wisp_routing.go`,
`labels.go`, `comments.go`, `dependency_queries.go`). Most upstream churn
won't collide. When it does:

1. Resolve conflicts in the working tree, keeping the spirit of our patch
   (single batched query instead of per-ID loops). If upstream refactored
   `IsActiveWispInTx` or its callers, re-port `PartitionWispIDsInTx`
   against the new shape.
2. **Re-run the quick validation:**
   ```bash
   cd ~/dev/beads
   go build -tags gms_pure_go ./internal/storage/issueops/
   make build
   time ./bd list --json > /dev/null     # sanity: should be ~5 s, not 25 s+
   ```
   If the timing regresses, upstream probably reintroduced an N+1 pattern
   somewhere new (check `cmd/bd/list.go` and the issueops helpers).
3. `git rebase --continue`, push, reinstall.

## When to re-sync

No schedule needed. Pull upstream when:

- A feature or fix we care about lands there (watch the upstream `CHANGELOG.md`).
- Our `bd` misbehaves in a way that might already be fixed upstream.
- Before cutting any new local patch branch, so we branch from fresh `main`.

## Submitting patches upstream

Our patches are not proprietary — most are general-interest perf and
bug fixes. When a branch stabilizes, open a PR from `Vonshlovens/beads` to
`gastownhall/beads`:

```bash
gh pr create --repo gastownhall/beads \
    --head Vonshlovens:<branch> \
    --base main \
    --title "..." \
    --body-file <description>
```

Once merged upstream, delete the local branch and rely on the synced `main`.
This keeps our fork as thin as possible.

## Current local patches

| Branch | Purpose | Doc | Upstream PR |
|---|---|---|---|
| `perf/bulk-wisp-partition` | 5× speedup on `bd list --json` over remote Dolt | [PERFORMANCE_FIX.md](./PERFORMANCE_FIX.md) | not yet opened |

When adding a new patch, append a row here.

## Installing on additional machines

Other machines just need the binary built from the same commit. Either:

- **Rebuild locally** (recommended — matches host arch, no notarisation
  issues on macOS):
  ```bash
  git clone git@github.com:Vonshlovens/beads.git ~/dev/beads
  cd ~/dev/beads && git checkout perf/bulk-wisp-partition
  make build && cp ./bd ~/.local/bin/bd
  ```
- **Or publish a release** in the fork (`gh release create`) and download
  the binary. We haven't bothered — rebuild is fast.
