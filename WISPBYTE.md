# DMG Browser Dashboard on Wispbyte

This branch builds DMG as a browser-first control service. The Go backend runs the existing grinder/Discord logic; the Svelte frontend is only a remote control UI.

## What changed

- Browser dashboard with login, overview, live logs, accounts, commands and settings.
- Existing auto-grind command logic is unchanged.
- Existing command defaults and command-specific settings remain in the Go/config code.
- Discord/account tokens are never returned by the browser API.
- Management API endpoints require an authenticated dashboard session.
- SSE live events are authenticated.
- Instance URLs use a short server-side ID instead of exposing the Discord token.
- The dashboard can start, stop and restart instances remotely.
- Wispbyte can keep the compiled Linux binary running continuously.

## Build the Linux web binary

The repository has a GitHub Actions workflow named **Web dashboard build**. It:

1. installs the frontend dependencies;
2. runs `npm run check`;
3. builds `frontend/dist`;
4. runs `go test ./...`;
5. embeds the generated frontend into the Go binary;
6. produces the `dmg-web-linux-amd64` artifact.

Download that artifact from the workflow and upload the `dmg-web` binary to Wispbyte.

Do not build only the Go binary before building the frontend: `main.go` embeds `frontend/dist`.

## Wispbyte setup

Wispbyte documents always-on hosting for Discord bots and web applications. Use a server/runtime that can execute the prebuilt Linux binary. Their panel provides Startup settings and environment variables.

Upload:

- `dmg-web`
- `config.json` if you use one

Recommended Startup command:

`chmod +x dmg-web && ./dmg-web`

The application listens on `PORT` when Wispbyte provides it, otherwise it uses port `5000`.

Health endpoint:

`/health`

## Required secrets

Set these in Wispbyte Startup → Environment Variables / Secrets:

| Name | Required | Purpose |
|---|---|---|
| `DASHBOARD_PASSWORD` | Yes | Password for the remote control dashboard |
| `DISCORD_TOKEN` or `TOKEN` | Optional | Existing environment-secret account support |
| `CHANNEL_ID` | Optional | Channel for the environment-secret account |
| `API_KEY` | Optional | Existing captcha/API configuration |

Keep Discord tokens and API keys in Wispbyte secrets instead of committing them to Git.

If `DASHBOARD_PASSWORD` is omitted, DMG generates a temporary password and prints a warning in the server console. Set the secret explicitly for normal use.

## Accessing the dashboard

Open the public web address assigned by Wispbyte. You will see the DMG login page.

After signing in:

- **Overview** — instances and live logs
- **Auto Grind** — the existing command settings
- **Accounts** — account state and instance controls
- **Settings** — cooldowns, breaks, events, auto-buy and auto-use

Press **Save changes** after changing configuration.

## Important hosting limitation

Wispbyte currently documents a one-Discord-bot-per-server policy on its free hosting. Check the current Wispbyte rules and your specific plan before running multiple Discord connections on one server.

The dashboard itself can be hosted as a web application, but Wispbyte's hosting policy and Discord's own rules still apply to the underlying workload.

## Security

Do not put `DASHBOARD_PASSWORD`, Discord tokens or API keys in the Git repository.

The browser API intentionally redacts:

- Discord account tokens
- the DMG API key

The backend restores protected values when saving normal dashboard configuration, so changing command/settings fields does not erase the stored token.

For production, access the dashboard over the HTTPS URL supplied by the host.
