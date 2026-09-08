# Run DMG 24/7 on Wispbyte (fishing-only)

This project is a **Go** app. Wispbyte free tier works best if you upload a **prebuilt Linux binary** (`dmg-web`) so you do not need a Go Docker image.

Free tier notes (as of 2026):
- ~512 MB RAM / ~1 GB disk / limited CPU
- Log in to [wispbyte.com/client](https://wispbyte.com/client) about every **2 weeks**
- **One Discord connection per server** (do not stack many accounts on free tier)

## 1. Prepare config locally

```bash
cp config.example.json config.json
```

Edit `config.json`:
- Set `state` to `true`
- Under `accounts`, add your Discord **user token** and **channel ID**
- Fishing is already enabled (`commands.fish.state` + `fishOnly`)

Never commit real tokens to GitHub.

## 2. Get a Linux binary

On a Linux PC (or WSL):

```bash
git clone https://github.com/aditya95595/dankeradimain.git
cd dankeradimain
go build -mod=vendor -o dmg-web .
```

Or use the repo’s existing `dmg-web` if it is already a Linux amd64 build.

## 3. Create a Wispbyte server

1. Sign in at https://wispbyte.com/client  
2. **Create Server** → Free Plan  
3. Pick any available image (Node/Python is fine if you only run the binary)  
4. Open **Files** and upload:
   - `dmg-web` (Linux binary)
   - `config.json` (with your token)
   - `start-wispbyte.sh` (optional)
   - `frontend/dist/` is already embedded in a full rebuild; binary alone is enough if built with embed

## 4. Startup settings

Open **Startup** and set:

**Startup command:**

```bash
chmod +x dmg-web start-wispbyte.sh 2>/dev/null; ./start-wispbyte.sh
```

or simply:

```bash
chmod +x dmg-web && ./dmg-web
```

The app reads **`PORT`** automatically (Wispbyte sets this). Health check: `/health`.

## 5. Start and verify

1. Click **Start** in Console  
2. Watch logs for `Logged in as ...` and fishing activity  
3. Open the dashboard URL shown by Wispbyte (host:port) if you want the simple UI  

## 24/7 tips

- Use **one account** on free tier (RAM/CPU limits).
- Keep `fishOnly: true` so other commands stay off.
- Optional breaks in config reduce ban risk; leave them on for safer long runs.
- Log into the Wispbyte panel every ~14 days on free tier.
- Private channel only; selfbots violate Discord ToS — use at your own risk.

## If it crashes

| Symptom | Fix |
|--------|-----|
| `Permission denied` | `chmod +x dmg-web` |
| Wrong architecture | Rebuild on Linux amd64 |
| Port in use / not binding | Ensure binary includes PORT support (latest `web_server.go`) |
| Invalid token | Fix token in `config.json` |
| OOM / killed | Use 1 account only; upgrade plan if needed |
