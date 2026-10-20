#!/bin/sh
# One-command rbx-sync install for Linux/WSL (linux amd64 asset).
set -eu

DRY_RUN=0
for arg in "$@"; do
	case "$arg" in
		--dry-run)
			DRY_RUN=1
			;;
		-h|--help)
			echo "usage: install.sh [--dry-run]"
			echo "Honor RBX_SYNC_RELEASE_BASE as the directory containing release assets (no trailing-slash required)."
			exit 0
			;;
	esac
done

ASSET_BIN="rbx-sync-linux-amd64"
ASSET_PLUGIN="rbx-sync.rbxmx"
ASSET_SUMS="SHA256SUMS"
DEST_BIN="${HOME}/.local/bin/rbx-sync"
DEST_PLUGIN="${HOME}/.local/share/rbx-sync/rbx-sync.rbxmx"

print_table() {
	echo "asset: ${ASSET_BIN}"
	echo "asset: ${ASSET_PLUGIN}"
	echo "asset: ${ASSET_SUMS}"
	echo "dest: ${DEST_BIN}"
	echo "dest: ${DEST_PLUGIN}"
	echo 'export PATH="$HOME/.local/bin:$PATH"'
}

if [ "${DRY_RUN}" -eq 1 ]; then
	print_table
	exit 0
fi

if [ -z "${RBX_SYNC_RELEASE_BASE:-}" ]; then
	echo "RBX_SYNC_RELEASE_BASE is unset; set it to the HTTPS directory of release assets (see README.md)." >&2
	exit 1
fi

BASE="${RBX_SYNC_RELEASE_BASE%/}"

file_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		openssl dgst -sha256 "$1" | awk '{print $NF}'
	fi
}

expected_hash() {
	sums=$1
	name=$2
	awk -v n="$name" '
		$2 == n || $2 == ("*" n) { print tolower($1); exit }
	' "$sums"
}

TMPDIR_INSTALL=$(mktemp -d)
cleanup() {
	rm -rf "${TMPDIR_INSTALL}"
}
trap cleanup EXIT

fetch() {
	name=$1
	curl -fsSL "${BASE}/${name}" -o "${TMPDIR_INSTALL}/${name}"
}

fetch "${ASSET_SUMS}"
fetch "${ASSET_BIN}"
fetch "${ASSET_PLUGIN}"

verify() {
	name=$1
	want=$(expected_hash "${TMPDIR_INSTALL}/${ASSET_SUMS}" "$name")
	got=$(file_sha256 "${TMPDIR_INSTALL}/${name}")
	got=$(echo "$got" | tr 'A-F' 'a-f')
	if [ -z "$want" ] || [ "$want" != "$got" ]; then
		echo "SHA-256 mismatch for ${name}; refusing to replace installed files." >&2
		exit 1
	fi
}

verify "${ASSET_BIN}"
verify "${ASSET_PLUGIN}"

mkdir -p "$(dirname "${DEST_BIN}")" "$(dirname "${DEST_PLUGIN}")"
cp "${TMPDIR_INSTALL}/${ASSET_BIN}" "${DEST_BIN}"
cp "${TMPDIR_INSTALL}/${ASSET_PLUGIN}" "${DEST_PLUGIN}"
chmod +x "${DEST_BIN}"

case ":${PATH}:" in
	*:"${HOME}/.local/bin":*) ;;
	*)
		echo 'export PATH="$HOME/.local/bin:$PATH"'
		;;
esac

plugins_reachable() {
	if command -v cmd.exe >/dev/null 2>&1; then
		win=$(cmd.exe /c "echo %LOCALAPPDATA%" 2>/dev/null | tr -d '\r')
		if [ -n "$win" ] && command -v wslpath >/dev/null 2>&1; then
			unix=$(wslpath -u "$win" 2>/dev/null || true)
			unix=$(echo "$unix" | tr -d '\r')
			if [ -n "$unix" ] && [ -d "${unix}/Roblox/Plugins" ]; then
				return 0
			fi
		fi
	fi
	n=0
	for d in /mnt/c/Users/*/AppData/Local/Roblox/Plugins; do
		if [ -d "$d" ]; then
			n=$((n + 1))
		fi
	done
	if [ "$n" -eq 1 ]; then
		return 0
	fi
	return 1
}

if plugins_reachable; then
	set +e
	"${DEST_BIN}" plugin install
	status=$?
	set -e
	exit "$status"
fi

echo "staged plugin: ${DEST_PLUGIN}"
exit 0
