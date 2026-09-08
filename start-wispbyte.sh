#!/usr/bin/env bash
set -euo pipefail

# Wispbyte 24/7 startup script for DMG (fishing-only)
# Preferred: run the prebuilt Linux binary if present.
# Fallback: build from source if Go is available.

cd "$(dirname "$0")"

if [ ! -f config.json ]; then
  if [ -f config.example.json ]; then
    cp config.example.json config.json
    echo "Created config.json from example. Edit accounts (token + channelID) then restart."
  else
    echo "Missing config.json — upload config with your Discord token and channel ID."
    exit 1
  fi
fi

# Make binary executable if present
if [ -f ./dmg-web ]; then
  chmod +x ./dmg-web
  exec ./dmg-web
fi

if command -v go >/dev/null 2>&1; then
  export GOTOOLCHAIN=local
  if [ -d vendor ]; then
    go build -mod=vendor -o ./dmg-web .
  else
    go build -o ./dmg-web .
  fi
  chmod +x ./dmg-web
  exec ./dmg-web
fi

echo "No dmg-web binary and no Go compiler found."
echo "Upload a Linux amd64 binary named dmg-web, or use a runtime with Go installed."
exit 1
