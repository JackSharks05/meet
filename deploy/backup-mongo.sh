#!/usr/bin/env bash
# Added by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
#
# Dumps the meet MongoDB to a timestamped, gzipped archive on the host, verifies
# the archive, optionally backs up deploy/.env (the secrets needed to actually
# restore the service), prunes old backups, and optionally copies offsite via
# rclone. Mongo is not published to the host, so we stream `mongodump` out of the
# container.
#
# Config via env (all optional):
#   MEET_BACKUP_DIR            where to write archives   (default: /home/jdh/meet-backups)
#   MEET_BACKUP_KEEP_DAYS      prune archives older than (default: 14)
#   MONGO_DB                   database name             (default: meet)
#   MEET_BACKUP_RCLONE_REMOTE  e.g. "gdrive:meet-backups" — copies backups offsite
#   MEET_BACKUP_GPG_PASSPHRASE if set, deploy/.env is encrypted to .env.gpg and
#                              ONLY the encrypted form is ever written / pushed
#                              offsite. Without it, .env is NOT copied offsite
#                              (plaintext secrets must not leave the box).
#
# Restore: see DEPLOY.md ("Restoring"). The .env backup is restored separately.
set -euo pipefail

BACKUP_DIR="${MEET_BACKUP_DIR:-/home/jdh/meet-backups}"
KEEP_DAYS="${MEET_BACKUP_KEEP_DAYS:-14}"
DB="${MONGO_DB:-meet}"
COMPOSE_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE="docker compose -f $COMPOSE_DIR/docker-compose.yml"

mkdir -p "$BACKUP_DIR"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$BACKUP_DIR/meet-$STAMP.archive.gz"

# --- Mongo dump --------------------------------------------------------------
# Stream the dump from inside the mongo container to a host file. -T disables the
# pseudo-TTY so the binary archive isn't corrupted.
#
# Run it in an `if !` so a non-zero exit is HANDLED here rather than tripping
# `set -e` first (which would leave the empty $OUT the redirect already created).
if ! $COMPOSE exec -T mongo mongodump --archive --gzip --db "$DB" > "$OUT"; then
  echo "backup FAILED: mongodump errored (is mongo up?)" >&2
  rm -f "$OUT"
  exit 1
fi

# Verify the archive is non-empty AND a readable gzip stream — "not empty" alone
# doesn't prove a partial/corrupt dump isn't garbage.
if [ ! -s "$OUT" ]; then
  echo "backup FAILED: $OUT is empty" >&2
  rm -f "$OUT"
  exit 1
fi
if ! gzip -t "$OUT" 2>/dev/null; then
  echo "backup FAILED: $OUT is not a valid gzip archive" >&2
  rm -f "$OUT"
  exit 1
fi
echo "wrote $OUT ($(du -h "$OUT" | cut -f1))"

# --- Config (.env) backup ----------------------------------------------------
# The Mongo dump alone can't stand the service back up — the secrets live in
# deploy/.env. Keep a copy alongside the dump, encrypted if a passphrase is set.
ENV_SRC="$COMPOSE_DIR/.env"
ENV_OFFSITE=""
if [ -f "$ENV_SRC" ]; then
  if [ -n "${MEET_BACKUP_GPG_PASSPHRASE:-}" ]; then
    ENV_OUT="$BACKUP_DIR/meet-env-$STAMP.env.gpg"
    if gpg --batch --yes --passphrase "$MEET_BACKUP_GPG_PASSPHRASE" \
         --symmetric --cipher-algo AES256 -o "$ENV_OUT" "$ENV_SRC"; then
      chmod 600 "$ENV_OUT"
      ENV_OFFSITE="$ENV_OUT" # encrypted → safe to send offsite
      echo "wrote $ENV_OUT (encrypted)"
    else
      echo "WARNING: failed to encrypt .env; skipping config backup" >&2
    fi
  else
    # No passphrase: keep a LOCAL plaintext copy only (never pushed offsite).
    ENV_OUT="$BACKUP_DIR/meet-env-$STAMP.env"
    cp "$ENV_SRC" "$ENV_OUT"
    chmod 600 "$ENV_OUT"
    echo "wrote $ENV_OUT (plaintext, local only — set MEET_BACKUP_GPG_PASSPHRASE to include offsite)"
  fi
else
  echo "WARNING: $ENV_SRC not found — config not backed up" >&2
fi

# --- Prune old local backups -------------------------------------------------
find "$BACKUP_DIR" -name 'meet-*.archive.gz' -type f -mtime +"$KEEP_DAYS" -delete
find "$BACKUP_DIR" -name 'meet-env-*.env*' -type f -mtime +"$KEEP_DAYS" -delete

# --- Offsite copy ------------------------------------------------------------
# Copies the Mongo archive, and the .env backup ONLY when it's encrypted.
if [ -n "${MEET_BACKUP_RCLONE_REMOTE:-}" ]; then
  if ! rclone copy "$OUT" "$MEET_BACKUP_RCLONE_REMOTE"; then
    echo "backup FAILED: rclone copy of $OUT errored" >&2
    exit 1
  fi
  if [ -n "$ENV_OFFSITE" ]; then
    rclone copy "$ENV_OFFSITE" "$MEET_BACKUP_RCLONE_REMOTE" \
      || echo "WARNING: rclone copy of encrypted .env failed" >&2
  fi
  echo "copied offsite to $MEET_BACKUP_RCLONE_REMOTE"
fi

# Heartbeat: only written on a fully-verified backup. An external monitor can
# alert when this file goes stale — "is it doing its job?", not "did it run?".
date +%s > "$BACKUP_DIR/last-success"
echo "backup OK ($STAMP)"
