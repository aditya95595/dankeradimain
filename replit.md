# Dank Memer Grinder — Web Dashboard

## Overview
DMG is a self-bot dashboard for automating Discord's Dank Memer bot.
Originally a Wails v3 desktop app, it has been converted to a web dashboard
that runs 24/7 on Replit and is reachable via any browser.

## Running the app
The Replit workflow builds the frontend and starts the Go web server automatically.

Manually:
```bash
cd frontend && bun run build && cd ..
go build -o dmg-web .
./dmg-web          # starts on port 5000 (web mode is default)
```

## URLs
| Purpose | Path |
|---------|------|
| Dashboard | `/` |
| Health check (UptimeRobot) | `/health` |
| REST API | `/api/*` |
| SSE event stream | `/api/events` |

## UptimeRobot
Point UptimeRobot to `https://<your-replit-domain>/health`.
It returns `{"status":"ok"}` with HTTP 200.

## Architecture
- **Go backend** (`web_server.go`): `net/http` server on port 5000 serving REST + SSE.
- **SSE hub** (`web_event.go`): broadcasts real-time events to all connected browsers.
- **Svelte frontend** (`frontend/`): built to `frontend/dist/` and served as static files.
- **Wails shim** (`frontend/src/lib/wails-shim.ts`): replaces `@wailsio/runtime` with SSE + fetch.

## User preferences
- Do not change bot/automation logic (Discord instance management, command handlers, etc.).
- Keep web-server concerns isolated to `web_server.go` and `web_event.go`.
