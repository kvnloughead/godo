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

if [ ! -d web/node_modules ]; then
	echo 'web/node_modules is missing. Run make web/install first.'
	exit 1
fi

echo '==> web: prettier'
npm --prefix web run --silent format:check

echo '==> web: typecheck and build'
npm --prefix web run --silent build

echo '==> web: test'
npm --prefix web test --silent

# The build above produced web/dist, so the embedded variant can be checked.
echo '==> go vet and build (-tags embedweb)'
go vet -tags embedweb ./web ./cmd/api
go build -tags embedweb -o /dev/null ./cmd/api

echo 'All checks passed.'
