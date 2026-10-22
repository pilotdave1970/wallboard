// Jet's Wallboard for Windows.
//
// A Windows port of the Raspberry Pi helper (wallboard_server.py). One .exe that:
//   - serves the board (built into the .exe) at http://localhost:8088/
//   - keeps settings in %APPDATA%\JetsWallboard\settings.json
//   - fetches Google Calendar (iCal) feeds and ADS-B receiver data for the page
//   - stores the pictures added for the Clock screen (%APPDATA%\JetsWallboard\pictures)
//   - opens the board in its own Microsoft Edge (or Chrome) app window, and quits when it's closed
//
// Command line:
//
//	JetsWallboard.exe               normal window (maximized)
//	JetsWallboard.exe --fullscreen  full screen (F11 or Alt+F4 to leave)
//	JetsWallboard.exe --lan         also let other devices on the home network open the board
//	JetsWallboard.exe --port 9000   use another port
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed wallboard.html
var boardHTML []byte

const ua = "JetsWallboard/1.0 (personal wall display for Windows)"

var (
	dataDir, settingsPath, picturesDir string
	lock                               sync.Mutex
	view                               = struct {
		View string `json:"view"`
		Seq  int    `json:"seq"`
	}{}
	viewLock sync.Mutex
	cache    = map[string]cached{}
	cacheMu  sync.Mutex
	client   = &http.Client{}
)

type cached struct {
	at     time.Time
	status int
	body   []byte
	ctype  string
}

var pictureTypes = []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg"}

func isPicture(name string) bool {
	l := strings.ToLower(name)
	for _, e := range pictureTypes {
		if strings.HasSuffix(l, e) {
			return true
		}
	}
	return false
}

// ---------- settings ----------

func readSettings() map[string]any {
	b, err := os.ReadFile(settingsPath)
	if err != nil {
		return map[string]any{}
	}
	var m map[string]any
	if json.Unmarshal(b, &m) != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func writeSettings(m map[string]any) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := settingsPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, settingsPath)
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9 _.-]+`)

func safePictureName(raw string) string {
	name, _ := url.QueryUnescape(raw)
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimSpace(name)
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.Trim(unsafeChars.ReplaceAllString(strings.TrimSuffix(name, filepath.Ext(name)), ""), " .")
	if len(base) > 60 {
		base = base[:60]
	}
	if base == "" {
		base = "picture"
	}
	if !isPicture(ext) || ext == "" {
		return ""
	}
	return base + ext
}

// ---------- fetching with a small cache ----------

func fetch(u string, maxAge, timeout time.Duration) (int, []byte, string) {
	now := time.Now()
	cacheMu.Lock()
	hit, ok := cache[u]
	cacheMu.Unlock()
	if ok && now.Sub(hit.at) < maxAge {
		return hit.status, hit.body, hit.ctype
	}
	out := cached{at: now}
	req, _ := http.NewRequest("GET", u, nil)
	if req == nil {
		out.status, out.body, out.ctype = 502, []byte("Bad address."), "text/plain"
	} else {
		req.Header.Set("User-Agent", ua)
		c := *client
		c.Timeout = timeout
		resp, err := c.Do(req)
		if err != nil {
			out.status, out.body, out.ctype = 502, []byte("Couldn't reach it: "+err.Error()), "text/plain"
		} else {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				out.status, out.body, out.ctype = 502, []byte(fmt.Sprintf("The server answered %d", resp.StatusCode)), "text/plain"
			} else {
				ct := resp.Header.Get("Content-Type")
				if ct == "" {
					ct = "application/octet-stream"
				}
				out.status, out.body, out.ctype = 200, body, ct
			}
		}
	}
	if out.status == 200 || !ok {
		cacheMu.Lock()
		cache[u] = out
		cacheMu.Unlock()
		return out.status, out.body, out.ctype
	}
	return hit.status, hit.body, hit.ctype // keep serving the last good copy while offline
}

// ---------- HTTP ----------

func send(w http.ResponseWriter, status int, body []byte, ctype string) {
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	w.Write(body)
}
func sendS(w http.ResponseWriter, status int, s, ctype string) { send(w, status, []byte(s), ctype) }
func sendJ(w http.ResponseWriter, status int, v any) {
	b, _ := json.Marshal(v)
	send(w, status, b, "application/json")
}

func lanIP() string {
	c, err := net.Dial("udp", "192.168.0.1:9")
	if err != nil {
		return ""
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}

var hexRe = regexp.MustCompile(`[^0-9a-f]`)
var csRe = regexp.MustCompile(`[^A-Z0-9]`)

func handler(lan bool, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		q := r.URL.Query()
		switch r.Method {
		case "GET", "HEAD":
			switch {
			case p == "/" || p == "/index.html" || p == "/wallboard.html":
				send(w, 200, boardHTML, "text/html; charset=utf-8")
			case p == "/api/settings":
				lock.Lock()
				s := readSettings()
				lock.Unlock()
				sendJ(w, 200, s)
			case p == "/api/calendar":
				var i int
				if _, err := fmt.Sscan(q.Get("i"), &i); err != nil {
					sendS(w, 404, "No calendar with that number.", "text/plain")
					return
				}
				u := ""
				if cals, ok := readSettings()["calendars"].([]any); ok && i >= 0 && i < len(cals) {
					if c, ok := cals[i].(map[string]any); ok {
						u, _ = c["url"].(string)
					}
				}
				u = strings.TrimSpace(u)
				if u == "" {
					sendS(w, 404, "No calendar with that number.", "text/plain")
					return
				}
				if strings.HasPrefix(u, "webcal://") {
					u = "https://" + u[len("webcal://"):]
				}
				if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
					sendS(w, 400, "Calendar address must start with https://", "text/plain")
					return
				}
				st, body, _ := fetch(u, 240*time.Second, 12*time.Second)
				ct := "text/plain"
				if st == 200 {
					ct = "text/calendar; charset=utf-8"
				}
				send(w, st, body, ct)
			case p == "/api/adsb":
				handleAdsb(w, r)
			case p == "/api/aircraftinfo":
				hx := hexRe.ReplaceAllString(strings.ToLower(q.Get("hex")), "")
				cs := csRe.ReplaceAllString(strings.ToUpper(q.Get("cs")), "")
				if len(hx) > 6 {
					hx = hx[:6]
				}
				if len(cs) > 8 {
					cs = cs[:8]
				}
				if len(hx) != 6 {
					sendS(w, 400, "Bad aircraft address.", "text/plain")
					return
				}
				u := "https://api.adsbdb.com/v0/aircraft/" + hx
				if cs != "" {
					u += "?callsign=" + cs
				}
				st, body, _ := fetch(u, 6*time.Hour, 8*time.Second)
				ct := "text/plain"
				if st == 200 {
					ct = "application/json"
				}
				send(w, st, body, ct)
			case p == "/api/photo":
				hx := hexRe.ReplaceAllString(strings.ToLower(q.Get("hex")), "")
				if len(hx) > 6 {
					hx = hx[:6]
				}
				if len(hx) != 6 {
					sendS(w, 400, "Bad aircraft address.", "text/plain")
					return
				}
				st, body, _ := fetch("https://api.planespotters.net/pub/photos/hex/"+hx, 24*time.Hour, 8*time.Second)
				ct := "text/plain"
				if st == 200 {
					ct = "application/json"
				}
				send(w, st, body, ct)
			case p == "/api/pc":
				handlePC(w, r, lan)
			case p == "/api/update/check":
				handleUpdateCheck(w, r)
			case p == "/api/view":
				viewLock.Lock()
				v := view
				viewLock.Unlock()
				sendJ(w, 200, v)
			case p == "/api/ping":
				host, _ := os.Hostname()
				ip := ""
				if lan {
					ip = lanIP()
				}
				sendJ(w, 200, map[string]any{"ok": true, "host": host, "ip": ip, "windows": true, "version": Version, "lan": lan, "port": port, "dir": dataDir})
			case strings.HasPrefix(p, "/pictures/"):
				name := safePictureName(strings.TrimPrefix(p, "/pictures/"))
				if name == "" {
					sendS(w, 404, "Not found", "text/plain")
					return
				}
				w.Header().Set("Cache-Control", "no-store")
				if strings.HasSuffix(name, ".svg") {
					w.Header().Set("Content-Type", "image/svg+xml")
				}
				http.ServeFile(w, r, filepath.Join(picturesDir, name))
			default:
				sendS(w, 404, "Not found", "text/plain")
			}
		case "DELETE":
			if p != "/api/pictures" {
				sendS(w, 404, "Not found", "text/plain")
				return
			}
			if name := safePictureName(q.Get("name")); name != "" {
				os.Remove(filepath.Join(picturesDir, name))
			}
			sendS(w, 200, `{"ok":true}`, "application/json")
		case "POST":
			switch p {
			case "/api/pc":
				handlePC(w, r, lan)
			case "/api/update/install":
				handleUpdateInstall(w, r)
			case "/api/hidepointer":
				// On the Pi this pressed a desktop key to hide the pointer. On Windows the page hides it itself.
				sendJ(w, 200, map[string]any{"ok": true, "detail": "handled by the page on Windows"})
			case "/api/pictures":
				name := safePictureName(q.Get("name"))
				if name == "" {
					sendS(w, 400, "Only PNG, SVG, JPG, WebP or GIF pictures can be used.", "text/plain")
					return
				}
				if r.ContentLength <= 0 || r.ContentLength > 25<<20 {
					sendS(w, 400, "That picture is too big (25 MB at most).", "text/plain")
					return
				}
				data, err := io.ReadAll(io.LimitReader(r.Body, 25<<20))
				if err != nil {
					sendS(w, 400, "Couldn't read the picture.", "text/plain")
					return
				}
				os.MkdirAll(picturesDir, 0o755)
				ext := filepath.Ext(name)
				base := strings.TrimSuffix(name, ext)
				final := name
				for i := 2; ; i++ {
					if _, err := os.Stat(filepath.Join(picturesDir, final)); errors.Is(err, os.ErrNotExist) {
						break
					}
					final = fmt.Sprintf("%s-%d%s", base, i, ext)
				}
				if err := os.WriteFile(filepath.Join(picturesDir, final), data, 0o644); err != nil {
					sendS(w, 500, "Couldn't save the picture.", "text/plain")
					return
				}
				sendJ(w, 200, map[string]string{"name": final})
			case "/api/view":
				var body struct {
					View string `json:"view"`
				}
				json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&body)
				switch body.View {
				case "home", "flights", "weather", "week", "bigclock":
				default:
					sendS(w, 400, "Unknown screen.", "text/plain")
					return
				}
				viewLock.Lock()
				view.View, view.Seq = body.View, view.Seq+1
				v := view
				viewLock.Unlock()
				sendJ(w, 200, v)
			case "/api/settings":
				var patch map[string]any
				b, err := io.ReadAll(io.LimitReader(r.Body, 2<<20+1))
				if err != nil || len(b) > 2<<20 {
					sendS(w, 400, "Settings must be a JSON object.", "text/plain")
					return
				}
				if len(strings.TrimSpace(string(b))) == 0 {
					b = []byte("{}")
				}
				if json.Unmarshal(b, &patch) != nil || patch == nil {
					sendS(w, 400, "Settings must be a JSON object.", "text/plain")
					return
				}
				lock.Lock()
				data := readSettings()
				for k, v := range patch {
					if v == nil {
						delete(data, k)
					} else {
						data[k] = v
					}
				}
				err = writeSettings(data)
				lock.Unlock()
				if err != nil {
					sendS(w, 500, "Couldn't save the settings file: "+err.Error(), "text/plain")
					return
				}
				sendJ(w, 200, data)
			default:
				sendS(w, 404, "Not found", "text/plain")
			}
		default:
			sendS(w, 405, "Not allowed", "text/plain")
		}
	}
}

// ---------- Windows bits ----------

func findBrowser() string {
	env := func(k string) string { return os.Getenv(k) }
	cands := []string{
		filepath.Join(env("ProgramFiles(x86)"), `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(env("ProgramFiles"), `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(env("LOCALAPPDATA"), `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(env("ProgramFiles"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(env("ProgramFiles(x86)"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(env("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
	}
	for _, c := range cands {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	if p, err := exec.LookPath("msedge.exe"); err == nil {
		return p
	}
	if p, err := exec.LookPath("chrome.exe"); err == nil {
		return p
	}
	return ""
}

// openBoard opens the board in its own app window and waits until that window is closed.
// Returns false if no Edge/Chrome was found (the board then opens in the default browser).
func openBoard(u string, fullscreen bool) bool {
	br := findBrowser()
	if br == "" {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
		return false
	}
	profile := filepath.Join(dataDir, "browser")
	// Clear the crash flag so a power cut never shows a "Restore pages?" bar.
	pref := filepath.Join(profile, "Default", "Preferences")
	if b, err := os.ReadFile(pref); err == nil {
		s := strings.ReplaceAll(string(b), `"exited_cleanly":false`, `"exited_cleanly":true`)
		s = regexp.MustCompile(`"exit_type":"[^"]*"`).ReplaceAllString(s, `"exit_type":"Normal"`)
		os.WriteFile(pref, []byte(s), 0o644)
	}
	args := []string{
		"--app=" + u,
		"--user-data-dir=" + profile,
		"--no-first-run", "--no-default-browser-check", "--noerrdialogs",
		"--disable-session-crashed-bubble", "--disable-features=Translate,TranslateUI",
		"--disable-background-mode", "--autoplay-policy=no-user-gesture-required",
		"--password-store=basic",
	}
	if fullscreen {
		args = append(args, "--start-fullscreen")
	} else {
		args = append(args, "--start-maximized")
	}
	cmd := exec.Command(br, args...)
	if err := cmd.Start(); err != nil {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
		return false
	}
	browserPID = cmd.Process.Pid
	cmd.Wait()
	return true
}

func setupDataDir() error {
	base := os.Getenv("APPDATA")
	if base == "" {
		base, _ = os.UserConfigDir()
	}
	dataDir = filepath.Join(base, "JetsWallboard")
	settingsPath = filepath.Join(dataDir, "settings.json")
	picturesDir = filepath.Join(dataDir, "pictures")
	if err := os.MkdirAll(picturesDir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(settingsPath); errors.Is(err, os.ErrNotExist) {
		// Bring over a settings.json copied from the Pi, if it sits next to the .exe.
		if exe, err := os.Executable(); err == nil {
			if b, err := os.ReadFile(filepath.Join(filepath.Dir(exe), "settings.json")); err == nil && json.Valid(b) {
				return os.WriteFile(settingsPath, b, 0o644)
			}
		}
		return os.WriteFile(settingsPath, []byte("{}"), 0o644)
	}
	return nil
}

var activePort int

// boardURL uses 127.0.0.1 rather than "localhost" so the browser never tries IPv6 (::1) first.
func boardURL(port int) string { return fmt.Sprintf("http://127.0.0.1:%d/", port) }

// isWallboard reports whether the program answering on this port is the wallboard itself
// (and not some other program that happens to use the same port).
func isWallboard(port int) bool {
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Get(boardURL(port) + "api/ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var j struct {
		OK      bool `json:"ok"`
		Windows bool `json:"windows"`
	}
	return resp.StatusCode == 200 && json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&j) == nil && j.OK && j.Windows
}

func main() {
	fullscreen := flag.Bool("fullscreen", false, "open full screen")
	lan := flag.Bool("lan", false, "let other devices on the home network open the board")
	port := flag.Int("port", 8088, "port")
	noWindow := flag.Bool("server-only", false, "run the helper without opening a window")
	pid := flag.Int("browser-pid", -1, "after an update: the board window to keep watching")
	flag.Parse()
	updated := *pid >= 0
	cleanupOld()

	if err := setupDataDir(); err != nil {
		messageBox("Jet's Wallboard", "Couldn't create the settings folder:\n"+err.Error())
		return
	}
	// Choices made in Settings > Board (the command-line switches still work too).
	if settingBool("pcFullscreen") {
		*fullscreen = true
	}
	if settingBool("pcLan") {
		*lan = true
	}
	host := "127.0.0.1"
	if *lan {
		host = "0.0.0.0"
	}
	// Find our port. If another program already uses it, move on to the next free one
	// (8088, 8089, ... 8107). If the wallboard itself is already running there, just open a window.
	var ln net.Listener
	var err error
	for try := 0; try < 20; try++ {
		p := *port + try
		ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, p))
		for i := 0; err != nil && updated && try == 0 && i < 60; i++ { // after an update, wait for the old version to let go
			time.Sleep(250 * time.Millisecond)
			ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, p))
		}
		if err == nil {
			*port = p
			break
		}
		if isWallboard(p) {
			if !*noWindow {
				openBoard(boardURL(p), *fullscreen)
			}
			return
		}
	}
	if err != nil {
		messageBox("Jet's Wallboard", fmt.Sprintf("Couldn't find a free port between %d and %d.\nStart it with --port 9000 (or another number) instead.", *port, *port+19))
		return
	}
	activePort = *port
	u := boardURL(*port)
	srv := &http.Server{Handler: handler(*lan, *port)}
	go srv.Serve(ln)
	restartFn = relaunch(func() { srv.Close() })
	quitFn = func() { srv.Close(); os.Exit(0) }

	if *noWindow {
		select {}
	}
	if updated {
		// Keep the window that was already open; quit when it closes.
		browserPID = *pid
		if *pid > 0 {
			p, err := os.FindProcess(*pid)
			if err != nil {
				openBoard(u, *fullscreen) // the window is gone; open a new one
				return
			}
			if _, err := p.Wait(); err != nil {
				for { // couldn't watch it directly; check now and then
					time.Sleep(5 * time.Second)
					if _, err := os.FindProcess(*pid); err != nil {
						return
					}
				}
			}
			return
		}
	} else if openBoard(u, *fullscreen) {
		return
	}
	messageBox("Jet's Wallboard", "Microsoft Edge or Google Chrome wasn't found, so the board opened in your default browser.\n\nKeep this message open while you use the board; click OK to close the wallboard helper.")
}
