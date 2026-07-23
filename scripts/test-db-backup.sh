#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test_directory="$(mktemp -d)"
fake_bin_directory="${test_directory}/bin"
backup_directory="${test_directory}/backups"

cleanup() {
  rm -rf -- "$test_directory"
}
trap cleanup EXIT

mkdir -p "$fake_bin_directory" "$backup_directory"

cat > "${fake_bin_directory}/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  inspect)
    if [ "${2:-}" = "-f" ]; then
      printf 'true\n'
    fi
    ;;
  exec)
    shift
    if [ "${1:-}" = "-i" ]; then
      shift
    fi
    shift
    case "${1:-}" in
      pg_dump)
        printf 'fake custom-format backup\n'
        ;;
      pg_restore)
        cat >/dev/null
        ;;
      *)
        echo "Unexpected docker exec command: ${1:-}" >&2
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unexpected docker command: ${1:-}" >&2
    exit 1
    ;;
esac
EOF
chmod +x "${fake_bin_directory}/docker"

touch "${backup_directory}/fungi_wiki-20000101T000000Z.dump"
touch "${backup_directory}/fungi_wiki-20010101T000000Z.dump"
touch "${backup_directory}/fungi_wiki-20020101T000000Z.dump"
touch "${backup_directory}/fungi_wiki-manual.dump"
touch "${backup_directory}/other_db-20000101T000000Z.dump"

new_backup="$(
  PATH="${fake_bin_directory}:${PATH}" \
    BACKUP_DIR="$backup_directory" \
    BACKUP_KEEP_COUNT=2 \
    "$repository_root/scripts/db-backup.sh"
)"

[ -s "$new_backup" ]
[ -f "${backup_directory}/fungi_wiki-20020101T000000Z.dump" ]
[ ! -e "${backup_directory}/fungi_wiki-20010101T000000Z.dump" ]
[ ! -e "${backup_directory}/fungi_wiki-20000101T000000Z.dump" ]
[ -f "${backup_directory}/fungi_wiki-manual.dump" ]
[ -f "${backup_directory}/other_db-20000101T000000Z.dump" ]

if PATH="${fake_bin_directory}:${PATH}" \
  BACKUP_DIR="$backup_directory" \
  BACKUP_KEEP_COUNT=invalid \
  "$repository_root/scripts/db-backup.sh" >/dev/null 2>&1; then
  echo "Invalid BACKUP_KEEP_COUNT unexpectedly succeeded" >&2
  exit 1
fi

echo "Database backup retention tests passed."
