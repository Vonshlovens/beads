# Beads `bd list --json` N+1 fix

## Problem

Against our Railway-hosted Dolt server over Tailscale (~70 ms RTT), `bd list
--json` took ~25–35 s on ~45 issues. Profiling showed the CLI was idle 99 %
of the time, waiting on serialized MySQL round trips.

## Root cause

Four bulk helpers in `internal/storage/issueops/` partitioned issue IDs into
wisp vs. permanent by calling `IsActiveWispInTx` **once per ID**:

- `GetLabelsForIssuesInTx` (`labels.go`)
- `GetCommentCountsInTx` (`comments.go`)
- `GetDependencyRecordsForIssuesInTx` (`dependency_queries.go`)
- `GetBlockingInfoForIssuesInTx` (`dependency_queries.go`)

`IsActiveWispInTx` runs `SELECT 1 FROM wisps WHERE id = ? LIMIT 1`. That's one
serial round trip per ID. `bd list --json` hits three of the four helpers
(`GetLabelsForIssuesInTx` actually runs twice — once via `SearchIssuesInTx`
hydration, once directly), so for 45 issues that's **~135–180 extra round
trips** just to figure out which IDs live in the wisps table.

## Fix

Added `PartitionWispIDsInTx` to `internal/storage/issueops/wisp_routing.go`:
a single batched `SELECT id FROM wisps WHERE id IN (…)` that returns
`(wispIDs, permIDs)` in one round trip. Each caller replaces its per-ID
loop with a single call to the new helper.

- Fork branch: `Vonshlovens/beads@perf/bulk-wisp-partition`
- Commit: `e3ad435c`
- Installed binary: `~/.local/bin/bd` (built locally with `make build`).

### Measured result

| Binary | `bd list --json` |
|---|---|
| upstream `1.0.2 (c446a2ef)` | ~26 s |
| patched (`perf/bulk-wisp-partition`) | ~5.5 s |

**≈5× faster.** Functionally identical; the legacy-DB fallback (missing
`wisps` table → treat all IDs as permanent) is preserved.

## Why this isn't hidden behind a config/env-var toggle

Worth considering — and rejected. The new path is strictly dominant:

- **No new dependencies, no schema change, no API change.** The partition
  helper uses the same `queryBatchSize` + `buildSQLInClause` machinery the
  four callers already used for their per-table queries.
- **Fallback path is preserved.** If `wisps` doesn't exist (legacy DBs
  pre-migration-007), `isTableNotExistError` returns all IDs as permanent
  — identical to the implicit `err == nil` fallback in `IsActiveWispInTx`.
- **The old code's "gracefully swallow errors" behaviour was wrong, not a
  feature.** `IsActiveWispInTx` returns `false` on *any* error (network
  blip, permission issue, etc.), silently mis-routing an ID. The new
  helper surfaces real errors. That's a small behavioural change, but in
  the direction of correctness.
- **A toggle would double the surface area to maintain** without any
  scenario where the old path is actually wanted.

**Counter-argument (why you might still want a toggle):** if we ever hit a
Dolt bug where the `IN (…)` clause misbehaves on a very large partition
(`> queryBatchSize`), a kill switch to revert to per-ID checks would be
useful. The existing `queryBatchSize = 200` cap plus the same pattern
already used elsewhere in the codebase makes this unlikely, but not
impossible.

**If we ever do add a toggle**, `BD_DISABLE_BULK_WISP_PARTITION=1` that
short-circuits `PartitionWispIDsInTx` back to a per-ID loop is the minimal
shape. We haven't done it because: YAGNI.

## Residual latency — what's still paid on every command

After the fix, `bd list --json` still spends ~5 s on the wire. Where it goes
(measured with one-off timing logs in `withReadTx`):

| Phase | Round trips | Approx. time |
|---|---|---|
| `openRoutedReadStore` → 6× `GetConfig` for routing keys | 6 × ~165 ms | ~1.0 s |
| `verifyProjectIdentity` | 1 | ~165 ms |
| `GetCustomStatuses` / `GetCustomStatusesDetailed` / `GetInfraTypes` | 3 | ~560 ms |
| `SearchIssues` (issues + wisps merge + hydrate labels) | 1 tx, 3 queries | ~800 ms |
| `GetLabelsForIssues`, `GetDependencyCounts`, `GetDependencyRecordsForIssues`, `GetCommentCounts` | 4 txs | ~1.2 s |
| Connection, handshake, misc | — | ~1.0 s |

## Future improvements (ordered by ROI)

### 1. Batch `getRoutingConfigValue` (easy, ~1 s win)

`determineAutoRoutedRepoPath` calls `store.GetConfig` up to 6 times for 6
different routing keys, each as its own `withReadTx`. Replace with a single
`GetConfigKeys(ctx, keys []string) map[string]string` that issues one
batched `IN` query. Single round trip instead of six.

### 2. Run the JSON-path auxiliary fetches concurrently (medium, ~0.5–1 s win)

In `cmd/bd/list.go` (~line 913 onward), `GetLabelsForIssues`,
`GetDependencyCounts`, `GetDependencyRecordsForIssues`, and
`GetCommentCounts` are issued serially. They're independent of each
other — an `errgroup.Group` fan-out would overlap ~4 × ~290 ms into
~400 ms. The underlying `withReadTx` already holds the store's RLock
re-entrantly, so parallelism is safe.

### 3. Skip the second `GetLabelsForIssues` call (easy, ~280 ms win)

`SearchIssuesInTx` already hydrates labels onto each issue. `list.go:913`
calls `GetLabelsForIssues` again anyway, overwriting with the same data.
Remove the second call for the JSON path.

### 4. Skip the wisps-merge pass in `SearchIssuesInTx` when infra types are excluded (medium, ~250 ms win)

When `filter.Ephemeral == nil` and the caller has already excluded all
infra types (which is the default for `bd list`), the second
`searchTableInTx(…, WispsFilterTables)` call is guaranteed to return zero
rows but still pays a round trip. Short-circuit it.

### 5. Batch the non-issues reads into one transaction (hard, ~0.5 s win)

Each of the 4 JSON-path auxiliary calls does its own BEGIN + SELECT +
ROLLBACK. Doing all four inside a single `withReadTx` amortizes two round
trips per call. Requires refactoring the `DoltStorage` interface or
adding a new `WithReadTx` export for `cmd/bd`.

Combined, #1–#4 would plausibly get `bd list --json` under 2 s without
touching network, and a sub-second time is realistic with #5 and
connection pooling tuning.

## Validating the fix locally

```bash
# From the beads clone at ~/dev/beads
make build              # builds ./bd
./bd list --json > /dev/null   # should return in ~5 s
cp ./bd ~/.local/bin/bd
```

## Upstreaming

The fix is a candidate for a PR against `gastownhall/beads`. See
[UPGRADE_WORKFLOW.md](./UPGRADE_WORKFLOW.md) for how we track upstream and
rebase our fork.
