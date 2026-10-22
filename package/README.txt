JET'S WALLBOARD FOR WINDOWS
===========================

Start it:  double-click JetsWallboard.exe
           (or "Jets Wallboard - Full Screen.cmd" to open it full screen; F11 or Alt+F4 to leave)

It opens the board in its own Microsoft Edge window. Close the window to quit.
No install and no Python needed. Windows 10 or 11.

FIRST RUN
  Windows SmartScreen may say "Windows protected your PC" because the .exe isn't signed.
  Click "More info" then "Run anyway". You only do this once.

SETTINGS
  Click the gear on the board, same as on the Pi. Everything is saved in:
      %APPDATA%\JetsWallboard   (settings.json and the pictures folder)
  To bring your Pi's settings over: copy settings.json from the Pi (~/wallboard/settings.json)
  into the same folder as JetsWallboard.exe BEFORE the first run. It gets picked up automatically.

AIRCRAFT (ADS-B)
  The board reads aircraft from your PiAware receiver. The default address is
      http://piaware.local/skyaware/data/aircraft.json
  If the Flights screen says "No receiver", open Settings and put in your receiver's IP, e.g.
      http://192.168.1.50/skyaware/data/aircraft.json
  and use the Test button.

START WITH WINDOWS (optional)
  Press Win+R, type  shell:startup  and press Enter. Put a shortcut to JetsWallboard.exe
  (or to the Full Screen .cmd) in that folder.

OPTIONS (add to a shortcut's Target)
  --fullscreen   open full screen
  --lan          let phones/tablets at home open the board at http://<this-PC>:8088
                 (Windows Firewall will ask once; allow Private networks)
  --port 9000    use another port if 8088 is taken

UPDATES
  Settings (gear) > Board > "Check for updates". If a newer version is out, click Install.
  It downloads the new version from GitHub, swaps it in and restarts in a few seconds;
  your settings and pictures are kept. Keep the JetsWallboard folder somewhere you can
  save files (Desktop, Documents) - not in Program Files - so it can update itself.

Source code and releases: https://github.com/pilotdave1970/wallboard
