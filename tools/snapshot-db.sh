#!/usr/bin/env bash
# Creates a timestamped snapshot of the SQLite database.
# Run from the repo root: ./tools/snapshot-db.sh
set -euo pipefail

# Locate the database file (matches os.UserCacheDir() on each platform)
case "$(uname -s)" in
  Darwin) CACHE_DIR="${HOME}/Library/Caches" ;;
  *)      CACHE_DIR="${XDG_CACHE_HOME:-${HOME}/.cache}" ;;
esac
DB_PATH="$CACHE_DIR/recipio/recipes.db"
SNAPSHOT_DIR="$CACHE_DIR/recipio/snapshots"

if [[ ! -f "$DB_PATH" ]]; then
  echo "Error: database not found at $DB_PATH" >&2
  exit 1
fi

mkdir -p "$SNAPSHOT_DIR"

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
DEST="$SNAPSHOT_DIR/recipes_${TIMESTAMP}.db"

cp "$DB_PATH" "$DEST"
echo "Snapshot saved: $DEST"
