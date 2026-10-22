# Jet's Wallboard

A wall display for your home PC: a big clock, weather with a radar map and National Weather Service warnings, live air traffic overhead, and your family calendar. It runs on Windows 10 or 11, and you don't need a Raspberry Pi or an antenna.

**Current version: v1.1.6** (October 21, 2026)

---

## Install

1. Download **JetsWallboard-v1.1.6.zip** from this page: click the file name, then **Download raw file** (the download arrow on the right).
2. Right-click the downloaded zip and choose **Extract All**. Put the folder somewhere you can save files, such as **Documents**. Don't use Program Files, because the app needs to update itself in that folder.
3. Open the folder and double-click **JetsWallboard.exe**.
4. The first time, Windows may say **"Windows protected your PC"**, because the app isn't code-signed. Click **More info**, then **Run anyway**.

The board opens in its own Microsoft Edge window. Close the window to quit.

You might like a shortcut: right-click **JetsWallboard.exe** and choose **Show more options → Send to → Desktop (create shortcut)**.

## First-time setup

The first time it runs, the board opens on **Settings → Location**:

1. Type your town in **Find a town** and click **Search**.
2. Pick your town from the list.
3. Click **Save location**.

That sets the weather, the radar map, the weather warnings and the center of the flight radar. Aircraft start showing up right away from a free online service.

## Using the board

The bar along the bottom switches screens:

| Button | What it shows |
|---|---|
| **Home** | Clock, weather, nearby aircraft and upcoming events together |
| **Flights** | Big flight radar and the list of nearby aircraft |
| **Weather** | Animated rain radar, active warnings, current conditions, the next 12 hours and a 7-day forecast |
| **Clock** | Just the time and date, large |
| **Week** | The family calendar for the week or month |

The small buttons on the right:

| Button | What it does |
|---|---|
| ⚙ Gear | Opens **Settings** |
| ▦ Layout | Moves and resizes the panels on Home |
| 🎨 Palette | Changes the color theme |
| ⛶ Corner arrows | **Full screen** on or off (you can also press **F**; **Esc** leaves) |
| ☾ Moon | Night dimming on or off |

**Aircraft:** tap any plane in the aircraft list to see its card: type, airline, route, altitude, speed and a photo from Planespotters.net. When a plane passes close overhead, the card pops up by itself.

## Settings (gear button)

**Location**
- Your town, for the weather, radar and warnings.
- **Center the flight radar on this place too?** Pick *No, somewhere else* to center it on a different spot, such as your house or your ADS-B receiver.

**Weather**
- Temperature units, radar map zoom, background map, loop length and speed, rain colors, and National Weather Service warnings with an optional banner.

**Calendar**
- **Google Calendar:**
  1. On a computer, open Google Calendar and go to **Settings**.
  2. Pick your calendar on the left and scroll to **Integrate calendar**.
  3. Copy the **Secret address in iCal format** and paste it into the board.
  4. Give it a name and choose who it's for.
- **Your own events:** add one-off or weekly events.
- **People and colors:** each person's events get their own color.

**Flights**
- **Where the aircraft come from:**
  - *Online (no receiver needed)*: pick **adsb.lol**, **airplanes.live** or **adsb.fi**. If one is slow, try another. Click **Test** to check it.
  - *My ADS-B receiver*: if you have one, enter its address, for example `http://192.168.1.50/skyaware/data/aircraft.json` or `.../tar1090/data/aircraft.json`. Click **Test** to check it.
- **Radar and list:** radar range, how many aircraft to list, colors, the map under the radar, airports, flight numbers, and whether to hide planes on the ground or above a set height.
- **Overhead alert:** turn it on or off, and set how close and how low a plane must be, how long the card stays up, whether to show photos, and whether to stay quiet at night.

**Display**
- Color theme, text colors, a picture behind the big clock, clock size and style, night dimming hours, the main screen, burn-in protection, and which screens appear in the bottom bar.

**Board**
- **This PC:**
  - **Start the wallboard when Windows starts**
  - **Open in full screen**
  - **Phones and tablets at home can open the board.** It shows the address to type on them.
- **Updates:** see below.
- **Restart the board** if something looks stuck. **Reset all settings** starts over.

## Updating

Each time the app starts, it checks this page for a newer version. If there is one, a box pops up on the board showing **what's new** in each version since yours:

- **Restart and update:** downloads the new version, restarts the app and reloads the board in a few seconds. Your settings are kept.
- **Later:** hides the box until the next time the app starts.

A board that's left running checks again every 6 hours. You can also check any time: gear → **Board** → **Check for updates**.

## Where your settings are kept

Settings and clock pictures are saved in `%APPDATA%\JetsWallboard`. To get there, paste that into the File Explorer address bar.

## Uninstall

Close the board and delete the JetsWallboard folder. If you turned on **Start the wallboard when Windows starts**, turn it off first. To remove your settings too, delete `%APPDATA%\JetsWallboard`.

## Troubleshooting

- **"Can't reach this page" or "refused to connect":** close every board window, wait 10 seconds and start **JetsWallboard.exe** again. Versions before v1.1.2 had this bug, so make sure you're on the latest version.
- **No aircraft ("offline"):** in **Settings → Flights**, pick a different online service and click **Test**.
- **No weather:** check that the PC is online. Weather comes from Open-Meteo and the National Weather Service.
- **Calendar not showing:** make sure you pasted the **Secret address in iCal format**, which ends in `.ics`, and not the calendar's web page address.

Aircraft data: [adsb.lol](https://adsb.lol), [airplanes.live](https://airplanes.live) and [adsb.fi](https://adsb.fi), for personal, non-commercial use. Aircraft details: [adsbdb](https://www.adsbdb.com). Photos: [Planespotters.net](https://www.planespotters.net). Weather: [Open-Meteo](https://open-meteo.com), [RainViewer](https://www.rainviewer.com) and the [National Weather Service](https://www.weather.gov).

---

## Revision history

| Version | Date | Changes |
|---|---|---|
| **v1.1.6** | October 21, 2026 | When the app starts it checks GitHub for an update. If there is one, a box shows what's new and offers **Restart and update** or **Later**. A board left running checks every 6 hours. |
| v1.1.5 | October 21, 2026 | Weather screen: the right side now shows current conditions (wind, gusts, humidity, rain chance, UV, pressure, visibility, sunrise and sunset), the next 12 hours and a 7-day forecast under the warnings. Moon icons at night. |
| v1.1.4 | October 21, 2026 | Check for updates now finds the newest zip on this GitHub page. |
| v1.1.3 | October 21, 2026 | Added the full-screen button in the bottom bar (or press F). |
| v1.1.2 | October 21, 2026 | Fixed "refused to connect": the board stays running while its window is open. |
| v1.1.1 | October 21, 2026 | Fixed "no page found" when another program uses port 8088: the board picks a free port. |
| v1.1.0 | October 21, 2026 | Works without a Raspberry Pi: online aircraft (adsb.lol, airplanes.live, adsb.fi), pick your home location on first run, tap a plane for its card and photo. Added Settings → Board → This PC: start with Windows, full screen, phones and tablets. |
| v1.0.0 | October 21, 2026 | First Windows version: the wallboard as a single JetsWallboard.exe, with Check for updates. |
