#!/usr/bin/env bash
# Runs the checks used by CI and the pre-commit hook. Run from anywhere in the
# repo: ./scripts/ci.sh or make ci.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo '==> gofmt'
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
	echo 'These files need gofmt -w:'
	echo "$unformatted"
	exit 1
fi

echo '==> go vet'
go vet ./...

echo '==> go build'
go build ./...

echo '==> go test'
go test -race ./...

echo 'All checks passed.'
