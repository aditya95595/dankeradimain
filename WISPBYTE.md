# Run DMG 24/7 on Wispbyte (fishing-only)

This project is a **Go** app. Wispbyte free tier works best if you upload a **prebuilt Linux binary** (`dmg-web`) so you do not need a Go Docker image.

Free tier notes (as of 2026):
- ~512 MB RAM / ~1 GB disk / limited CPU
- Log in to [wispbyte.com/client](https://wispbyte.com/client) about every **2 weeks**
- **One Discord connection per server** (do not stack many accounts on free tier)

## Secrets (recommended)

In Wispbyte **Startup → Environment variables / Secrets**, set:

| Name | Required | Description |
|------|----------|-------------|
| `DISCORD_TOKEN` | Yes | Your Discord **user** token (or use `TOKEN`) |
| `CHANNEL_ID` | Yes | Channel ID where Dank Memer is used |
| `API_KEY` | No | Captcha solver API key |
| `PORT` | Auto | Usually set by Wispbyte — do not override unless needed |

The bot reads these on startup and injects them into the first account. You can leave `accounts: []` empty in `config.json`.

## 1. Prepare config (no secrets in the file)

```bash
cp config.example.json config.json
```

Keep fishing enabled (`commands.fish.state` + `fishOnly`). Leave token/channel empty if using secrets.

## 2. Get a Linux binary

On a Linux PC (or WSL):

```bash
git clone https://github.com/aditya95595/dankeradimain.git
cd dankeradimain
go build -mod=vendor -o dmg-web .
```

Rebuild after pulling so env-secret support is included.

## 3. Create a Wispbyte server

1. Sign in at https://wispbyte.com/client  
2. **Create Server** → Free Plan  
3. Pick any available image (Node/Python is fine if you only run the binary)  
4. **Files** → upload `dmg-web` + `config.json` (+ optional `start-wispbyte.sh`)  
5. **Startup** → add secrets listed above  

## 4. Startup command

```bash
chmod +x dmg-web && ./dmg-web
```

or:

```bash
chmod +x dmg-web start-wispbyte.sh 2>/dev/null; ./start-wispbyte.sh
```

Health check path: `/health`

## 5. Start and verify

1. **Start** the server  
2. Console should log that secrets were applied, then `Logged in as ...`  
3. Fishing should begin in the configured channel  

## 24/7 tips

- One account on free tier  
- `fishOnly: true`  
- Log into the panel every ~14 days on free tier  
- Private channel only; selfbots violate Discord ToS — use at your own risk  

## Troubleshooting

| Symptom | Fix |
|--------|-----|
| Permission denied | `chmod +x dmg-web` |
| Wrong arch | Rebuild Linux amd64 |
| Invalid token | Check `DISCORD_TOKEN` / `TOKEN` secret |
| No channel | Set `CHANNEL_ID` secret |
| OOM | One account only; upgrade plan |
