# Jet's Wallboard

A wall display for home: clock, weather with radar and NWS warnings, live air traffic (online, or from your own ADS-B receiver), and the family's Google Calendar. Runs as a single Windows program, `JetsWallboard.exe`.

## Download

Get **JetsWallboardSetup.exe** from the [latest release](https://github.com/pilotdave1970/wallboard/releases/latest) and run it. It installs just for you (no admin password), adds Start menu and desktop shortcuts, and opens the board in its own Microsoft Edge window. Windows 10 or 11 only.

Windows may show "Windows protected your PC" the first time because the app isn't code-signed: click **More info → Run anyway**.

Everything else is set on the board itself (gear → **Board**):

- **This PC:** start with Windows, open full screen, let phones and tablets at home open the board.
- **Updates:** check for and install new versions.

Uninstall from Windows **Settings → Apps**. Your settings stay in `%APPDATA%\JetsWallboard`.

**No Raspberry Pi needed.** On first run it asks for your town. Live aircraft come from a free community ADS-B network (adsb.lol by default; airplanes.live and adsb.fi are also available), or from your own receiver if you have one. Tap a plane to see its type, route and a [Planespotters.net](https://www.planespotters.net) photo.

## Releasing a new version

1. Commit your changes to `main`.
2. On GitHub: **Releases → Draft a new release**, create a new tag that is higher than the last one (`v1.0.1`, `v1.1.0`, …), write a few lines about what changed, and **Publish**.
3. The *Build Windows release* action runs on a Windows machine at GitHub. It builds the app with that version number, wraps it in an installer with [Inno Setup](https://jrsoftware.org/isinfo.php) and attaches `JetsWallboardSetup.exe` to the release (a few minutes; watch the **Actions** tab).

Every running board then sees the new version under *Check for updates*, and the release notes are shown before installing.

## How it's built

| File | What it is |
|---|---|
| `wallboard.html` | The whole board (HTML/CSS/JS). Built into the .exe. |
| `main.go` | Local helper on `localhost:8088`: settings, Google Calendar and ADS-B fetching, pictures, opens the Edge app window. |
| `adsb.go` | Aircraft data: your own receiver, or an online ADS-B network around your radar center. |
| `update.go` | Checks GitHub's latest release, downloads `JetsWallboardSetup.exe` and runs it silently; the board window stays open. |
| `pc.go`, `pc_windows.go` | The **This PC** settings: start with Windows, full screen, phones and tablets. |
| `rsrc_windows_amd64.syso` | The icon and Windows manifest (made from `assets/` with [`rsrc`](https://github.com/akavel/rsrc)). |
| `installer/JetsWallboard.iss` | The installer script (Inno Setup). |

Build it yourself with Go 1.24+:

```
set GOOS=windows
go build -ldflags "-H windowsgui -X main.Version=v1.0.0" -o JetsWallboard.exe .
```
