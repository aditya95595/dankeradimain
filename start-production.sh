#!/usr/bin/env bash
set -euo pipefail

# The deployment image may not include locally generated binaries. Build from
# the tracked source at startup and execute from a writable location.
export GOTOOLCHAIN=local
go build -mod=vendor -o /tmp/dmg-web .
exec /tmp/dmg-web