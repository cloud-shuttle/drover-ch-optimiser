#!/usr/bin/env bash
# Cross-compile drover-ch and drover-ch-mcp (fallback when GoReleaser CI is unavailable).
set -euo pipefail
VERSION="${1:-0.1.0}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist"
rm -rf "$DIST"
mkdir -p "$DIST"

build_one() {
  local pkg="$1" bin="$2" goos="$3" goarch="$4"
  local out="$DIST/${bin}_${VERSION}_${goos}_${goarch}"
  [[ "$goos" == "windows" ]] && out="${out}.exe"
  echo "  $goos/$goarch -> $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "$out" "$ROOT/$pkg"
}

for goos in linux darwin windows; do
  for goarch in amd64 arm64; do
    build_one "./cmd/drover-ch" "drover-ch" "$goos" "$goarch"
    build_one "./cmd/drover-ch-mcp" "drover-ch-mcp" "$goos" "$goarch"
  done
done

(
  cd "$DIST"
  for f in drover-ch_* drover-ch-mcp_*; do
    [[ -f "$f" ]] || continue
    case "$f" in
      *.exe) zip -q "${f%.exe}.zip" "$f" ;;
      *) tar czf "${f}.tar.gz" "$f" ;;
    esac
  done
  if command -v sha256sum >/dev/null; then
    sha256sum ./*.tar.gz ./*.zip > checksums.txt
  else
    shasum -a 256 ./*.tar.gz ./*.zip > checksums.txt
  fi
)

echo "Artifacts in $DIST"
