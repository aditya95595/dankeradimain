> It is crucial that you take appropriate measures to avoid detection. These measures include:
> <ol>
>  <li>Running the bot only in private channels.</li>
>  <li>Not being open about the fact that you use it.</li>
>  <li>If you run the bot excessively, no matter the outcome of captchas. Big surprise, if you farm for 12 hours straight then they might think you're a little bit sussy. So if you script for an unnatural amount of time, then don't be surprised if you get banned for that very reason.</li>
> </ol>

## Current Focus

This fork is configured for **fishing-only** mode by default.

- Fishing is enabled and set to `fishOnly: true`
- Auto equipment buying is enabled
- All other commands are disabled by default

## Features
A detailed breakdown of features are available in the original [**<u>documentation</u>**](https://docs.dankmemer.tools/features/commands).

-   [x] **Undetectable:** Mimics native api's used by official discord clients, making it impossible for discord to tell the difference.
-   [x] **Accurate and Cheap Captcha Solver (Not public yet):** Automatically solve captchas on many accounts while being undetectable, only at $0.005 per captcha solved ($0.5 / 100 captchas).
-   [x] **Performant:** DMG is written in Go with a focus on speed, using minimal system resources in the background.
-   [x] **Highly Customizable:** Through DMG's modern GUI, everything from search priorities to adventure answers can be easily customized to your hearts content.
-   [x] **Fishing Support:** Full fishing automation including location selection, equipment, buckets, and catch minigame handling.
-   [x] **Multi-Platform Executables** No more installing python or node. Just download and run the executable file available for Windows, macOS and Linux.

<div align="center">
  <img src=".github/assets/img/logs.png" width="400">
  <img src=".github/assets/img/settings.png" width="400">
  <img src=".github/assets/img/accounts.png" width="400">
  <img src=".github/assets/img/commands.png" width="400">
</div>

## Quick Start (Fishing Only)

1. Copy the example config:
   ```bash
   cp config.example.json config.json
   ```

2. Add your Discord token and channel ID in `config.json` under `accounts`.

3. Run in web mode:
   ```bash
   go run . -web
   ```
   or use the prebuilt binary:
   ```bash
   ./dmg-web
   ```

4. Open the dashboard (default http://0.0.0.0:5000)

## Documentation

### Installation
- [Pre-Built binaries (Recommended)](https://docs.dankmemer.tools/installation/pre-built-binaries)
- [Build from source](https://docs.dankmemer.tools/installation/build-from-source)

### Getting Started
- [Entering token and channel ID](https://docs.dankmemer.tools/configuration/entering-token-and-channel-id)
- [config.json file](https://docs.dankmemer.tools/configuration/config-json)
- [General settings](https://docs.dankmemer.tools/configuration/general-settings)
- [Commands Settings](https://docs.dankmemer.tools/configuration/commands-settings)
- [Auto buy settings](https://docs.dankmemer.tools/configuration/auto-buy-settings)
- [Auto use settings](https://docs.dankmemer.tools/configuration/auto-use-settings)

### Features
- [Commands](https://docs.dankmemer.tools/features/commands)
- [Minigames](https://docs.dankmemer.tools/features/minigames)

## Credits
- [dgate](https://github.com/LuminalDev/dgate)
- [disgo](https://github.com/disgoorg/disgo)
- [catchtwo](https://github.com/kyan0045/CatchTwo)
- [slashy](https://github.com/TahaGorme/slashy)

## Star History
<div align="center">
  <a href="https://star-history.com/#bridgesensedev/dmg&Date">
    <img src="https://api.star-history.com/svg?repos=autocord-org/dmg&type=Date" 
         alt="Star History Chart">
  </a>
</div>
