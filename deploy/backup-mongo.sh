#!/usr/bin/env bash
# Added by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
#
# Dumps the meet MongoDB to a timestamped, gzipped archive on the host, prunes
# old backups, and optionally copies the latest offsite via rclone. Mongo is not
# published to the host, so we stream `mongodump` out of the container.
#
# Config via env (all optional):
#   MEET_BACKUP_DIR         where to write archives   (default: /home/jdh/meet-backups)
#   MEET_BACKUP_KEEP_DAYS   prune archives older than (default: 14)
#   MONGO_DB                database name             (default: meet)
#   MEET_BACKUP_RCLONE_REMOTE  e.g. "gdrive:meet-backups" — copies latest offsite
#
# Restore a backup with:
#   docker compose exec -T mongo mongorestore --archive --gzip --drop < FILE.archive.gz
set -euo pipefail

BACKUP_DIR="${MEET_BACKUP_DIR:-/home/jdh/meet-backups}"
KEEP_DAYS="${MEET_BACKUP_KEEP_DAYS:-14}"
DB="${MONGO_DB:-meet}"
COMPOSE_DIR="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$BACKUP_DIR"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$BACKUP_DIR/meet-$STAMP.archive.gz"

# Stream the dump from inside the mongo container to a host file. -T disables the
# pseudo-TTY so the binary archive isn't corrupted.
docker compose -f "$COMPOSE_DIR/docker-compose.yml" exec -T mongo \
  mongodump --archive --gzip --db "$DB" > "$OUT"

# Guard against a zero-byte dump (e.g. mongo not ready) — don't keep a bad file.
if [ ! -s "$OUT" ]; then
  echo "backup FAILED: $OUT is empty" >&2
  rm -f "$OUT"
  exit 1
fi
echo "wrote $OUT ($(du -h "$OUT" | cut -f1))"

# Prune old local archives.
find "$BACKUP_DIR" -name 'meet-*.archive.gz' -type f -mtime +"$KEEP_DAYS" -delete

# Optional offsite copy (needs rclone configured for the named remote).
if [ -n "${MEET_BACKUP_RCLONE_REMOTE:-}" ]; then
  rclone copy "$OUT" "$MEET_BACKUP_RCLONE_REMOTE" \
    && echo "copied offsite to $MEET_BACKUP_RCLONE_REMOTE"
fi
