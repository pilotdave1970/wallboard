# Jet's Wallboard

A wall display for home: clock, weather with radar and NWS warnings, live air traffic from your own ADS-B receiver, and the family's Google Calendar. Runs as a single Windows program, `JetsWallboard.exe`.

## Download

Get **JetsWallboard-Windows.zip** from the [latest release](https://github.com/pilotdave1970/wallboard/releases/latest), unzip it somewhere like Documents, and double-click `JetsWallboard.exe`. The board opens in its own Microsoft Edge window. Windows 10 or 11; nothing else to install.

**No Raspberry Pi needed.** On first run it asks for your town. Live aircraft come from a free community ADS-B network (adsb.lol by default; airplanes.live and adsb.fi are also available), or from your own receiver if you have one. Tap a plane to see its type, route and a [Planespotters.net](https://www.planespotters.net) photo.

Settings live in `%APPDATA%\JetsWallboard`. The app updates itself: **Settings → Board → Check for updates**.

Options for a shortcut: `--fullscreen`, `--lan` (open the board from other devices at home), `--port 9000`.

## Releasing a new version

1. Commit your changes to `main`.
2. On GitHub: **Releases → Draft a new release**, create a new tag that is higher than the last one (`v1.0.1`, `v1.1.0`, …), write a few lines about what changed, and **Publish**.
3. The *Build Windows release* action builds `JetsWallboard.exe` with that version number and attaches it plus the zip to the release (about a minute; watch the **Actions** tab).

Every running board then sees the new version under *Check for updates*, and the release notes are shown before installing.

## How it's built

| File | What it is |
|---|---|
| `wallboard.html` | The whole board (HTML/CSS/JS). Built into the .exe. |
| `main.go` | Local helper on `localhost:8088`: settings, Google Calendar and ADS-B fetching, pictures, opens the Edge app window. |
| `adsb.go` | Aircraft data: your own receiver, or an online ADS-B network around your radar center. |
| `update.go` | Checks GitHub's latest release, downloads `JetsWallboard.exe`, swaps it in and restarts without closing the board window. |
| `rsrc_windows_amd64.syso` | The icon and Windows manifest (made from `assets/` with [`rsrc`](https://github.com/akavel/rsrc)). |
| `package/` | Extra files that go in the download zip. |

Build it yourself with Go 1.24+:

```
set GOOS=windows
go build -ldflags "-H windowsgui -X main.Version=v1.0.0" -o JetsWallboard.exe .
```
