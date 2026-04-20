package issueops

import (
	"context"
	"database/sql"
	"fmt"
)

// IsActiveWispInTx checks whether the given ID exists in the wisps table
// within an existing transaction. Returns true if the ID is found.
// This handles both auto-generated wisp IDs (containing "-wisp-") and
// ephemeral issues created with explicit IDs that were routed to wisps.
func IsActiveWispInTx(ctx context.Context, tx *sql.Tx, id string) bool {
	var exists int
	err := tx.QueryRowContext(ctx, "SELECT 1 FROM wisps WHERE id = ? LIMIT 1", id).Scan(&exists)
	return err == nil
}

// PartitionWispIDsInTx splits issueIDs into (wispIDs, permIDs) with a single
// batched query against the wisps table, instead of N per-ID existence checks.
// This is a significant latency win when the caller is over a remote Dolt
// connection: N round trips collapse to ceil(N/queryBatchSize) round trips.
//
// If the wisps table does not exist (legacy databases pre-migration-007), all
// IDs are returned as permIDs — matching the IsActiveWispInTx fallback.
// Input order is preserved within each partition.
func PartitionWispIDsInTx(ctx context.Context, tx *sql.Tx, issueIDs []string) (wispIDs, permIDs []string, err error) {
	if len(issueIDs) == 0 {
		return nil, nil, nil
	}

	wispSet := make(map[string]struct{})
	for start := 0; start < len(issueIDs); start += queryBatchSize {
		end := start + queryBatchSize
		if end > len(issueIDs) {
			end = len(issueIDs)
		}
		batch := issueIDs[start:end]
		inClause, args := buildSQLInClause(batch)
		//nolint:gosec // G201: inClause is composed solely of ? placeholders
		rows, qerr := tx.QueryContext(ctx,
			fmt.Sprintf("SELECT id FROM wisps WHERE id IN (%s)", inClause), args...)
		if qerr != nil {
			// Legacy database without wisps table: treat all IDs as permanent.
			if isTableNotExistError(qerr) {
				return nil, append([]string(nil), issueIDs...), nil
			}
			return nil, nil, fmt.Errorf("partition wisp ids: %w", qerr)
		}
		for rows.Next() {
			var id string
			if scanErr := rows.Scan(&id); scanErr != nil {
				_ = rows.Close()
				return nil, nil, fmt.Errorf("partition wisp ids: scan: %w", scanErr)
			}
			wispSet[id] = struct{}{}
		}
		_ = rows.Close()
		if rerr := rows.Err(); rerr != nil {
			return nil, nil, fmt.Errorf("partition wisp ids: rows: %w", rerr)
		}
	}

	for _, id := range issueIDs {
		if _, ok := wispSet[id]; ok {
			wispIDs = append(wispIDs, id)
		} else {
			permIDs = append(permIDs, id)
		}
	}
	return wispIDs, permIDs, nil
}

// WispTableRouting returns the appropriate issue, label, event, and dependency
// table names based on whether the ID is an active wisp. Call IsActiveWispInTx
// first to determine isWisp.
func WispTableRouting(isWisp bool) (issueTable, labelTable, eventTable, depTable string) {
	if isWisp {
		return "wisps", "wisp_labels", "wisp_events", "wisp_dependencies"
	}
	return "issues", "labels", "events", "dependencies"
}
