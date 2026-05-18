#!/usr/bin/env sh
# Snapshots the recipio SQLite DB (excluding the sessions table).
# Skips if no meaningful data has changed since the last snapshot.
# Keeps at most MAX_SNAPSHOTS entries, pruning oldest first.
#
# Local dev usage: ./tools/snapshot-db.sh
# In Docker: invoked by crond every 6 hours via the entrypoint.
set -eu

# Resolve DB path: env var > os.UserCacheDir() convention
if [ -n "${RECIPIO_DB_PATH:-}" ]; then
  DB_PATH="$RECIPIO_DB_PATH"
else
  case "$(uname -s)" in
    Darwin) DB_PATH="${HOME}/Library/Caches/recipio/recipes.db" ;;
    *)      DB_PATH="${XDG_CACHE_HOME:-${HOME}/.cache}/recipio/recipes.db" ;;
  esac
fi

SNAPSHOT_DIR="$(dirname "$DB_PATH")/snapshots"
MAX_SNAPSHOTS=28
HASH_FILE="$SNAPSHOT_DIR/.last_hash"

if [ ! -f "$DB_PATH" ]; then
  echo "snapshot: db not found at $DB_PATH, skipping"
  exit 0
fi

mkdir -p "$SNAPSHOT_DIR"

# Hash the content of every table except sessions to detect meaningful changes.
TABLES=$(sqlite3 "$DB_PATH" \
  "SELECT name FROM sqlite_master WHERE type='table' AND name != 'sessions' ORDER BY name;")
CURRENT_HASH=$(
  for TABLE in $TABLES; do
    sqlite3 "$DB_PATH" "SELECT * FROM \"$TABLE\" ORDER BY rowid;"
  done | md5sum | cut -d' ' -f1
)

if [ -f "$HASH_FILE" ] && [ "$(cat "$HASH_FILE")" = "$CURRENT_HASH" ]; then
  echo "snapshot: no changes since last snapshot, skipping"
  exit 0
fi

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
DEST="$SNAPSHOT_DIR/recipes_${TIMESTAMP}.db"

# .backup uses SQLite's Online Backup API — safe with WAL mode.
sqlite3 "$DB_PATH" ".backup $DEST"
# Strip sessions data and reclaim space.
sqlite3 "$DEST" "DELETE FROM sessions; VACUUM;"

echo "$CURRENT_HASH" > "$HASH_FILE"

# Prune: keep only the MAX_SNAPSHOTS most recent files.
ls -1t "$SNAPSHOT_DIR"/recipes_*.db 2>/dev/null | tail -n +$((MAX_SNAPSHOTS + 1)) | while IFS= read -r OLD; do
  rm -- "$OLD"
  echo "snapshot: pruned $OLD"
done

echo "snapshot: saved $DEST"
