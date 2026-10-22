package main

// Self-update from GitHub Releases.
//
//   GET  /api/update/check    -> {current, latest, available, notes, url, ready}
//   POST /api/update/install  -> downloads the new JetsWallboard.exe, swaps it in, restarts the helper.
//
// Each release carries one file, JetsWallboardSetup.exe. Installing an update downloads it and
// runs it silently; the new version takes over the same port and keeps watching the same
// browser window, and the page reloads itself once the new version answers.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Version is set at build time: -ldflags "-X main.Version=v1.2.3"
var Version = "dev"

const repo = "pilotdave1970/wallboard"
const assetName = "JetsWallboardSetup.exe"

type release struct {
	Tag    string `json:"tag_name"`
	URL    string `json:"html_url"`
	Body   string `json:"body"`
	Draft  bool   `json:"draft"`
	Pre    bool   `json:"prerelease"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

var (
	updateMu   sync.Mutex
	installing bool
	exePath    string                      // where the program file lives, read once at startup (before any update renames it)
	browserPID int                         // the board window's browser process, handed to the next version
	restartFn  func(extra ...string) error // set in main
	quitFn     func()                      // set in main
)

func latestRelease() (*release, error) {
	api := "https://api.github.com/repos/" + repo + "/releases/latest"
	if t := os.Getenv("WALLBOARD_UPDATE_API"); t != "" { // for testing only
		api = t
	}
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't reach GitHub: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases have been published yet")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub answered %d", resp.StatusCode)
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("couldn't read GitHub's answer")
	}
	return &r, nil
}

func (r *release) asset() (string, int64) {
	for _, a := range r.Assets {
		if strings.EqualFold(a.Name, assetName) {
			return a.URL, a.Size
		}
	}
	return "", 0
}

// parseVer turns "v1.2.10" into [1 2 10]. Anything that isn't a number counts as 0.
func parseVer(s string) [3]int {
	var v [3]int
	s = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(s)), "v")
	if i := strings.IndexAny(s, "-+ "); i >= 0 {
		s = s[:i]
	}
	for i, p := range strings.SplitN(s, ".", 3) {
		v[i], _ = strconv.Atoi(p)
	}
	return v
}

func newer(latest, current string) bool {
	if current == "dev" {
		return true
	}
	a, b := parseVer(latest), parseVer(current)
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	rel, err := latestRelease()
	if err != nil {
		sendJ(w, 200, map[string]any{"current": Version, "error": err.Error()})
		return
	}
	u, _ := rel.asset()
	sendJ(w, 200, map[string]any{
		"current":   Version,
		"latest":    rel.Tag,
		"available": newer(rel.Tag, Version),
		"ready":     u != "", // false while GitHub is still building the .exe for a brand-new release
		"notes":     rel.Body,
		"url":       rel.URL,
	})
}

func handleUpdateInstall(w http.ResponseWriter, r *http.Request) {
	updateMu.Lock()
	if installing {
		updateMu.Unlock()
		sendS(w, 409, "An update is already being installed.", "text/plain")
		return
	}
	installing = true
	updateMu.Unlock()
	fail := func(status int, msg string) {
		updateMu.Lock()
		installing = false
		updateMu.Unlock()
		sendS(w, status, msg, "text/plain")
	}

	rel, err := latestRelease()
	if err != nil {
		fail(502, err.Error())
		return
	}
	if !newer(rel.Tag, Version) {
		fail(400, "You already have the latest version.")
		return
	}
	u, size := rel.asset()
	if u == "" {
		fail(409, "GitHub is still building "+rel.Tag+". Try again in a couple of minutes.")
		return
	}
	setup := filepath.Join(os.TempDir(), "JetsWallboardSetup-"+strings.TrimPrefix(rel.Tag, "v")+".exe")

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", ua)
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		fail(502, "Download failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fail(502, fmt.Sprintf("Download failed: GitHub answered %d", resp.StatusCode))
		return
	}
	f, err := os.Create(setup)
	if err != nil {
		fail(500, "Couldn't save the update: "+err.Error())
		return
	}
	n, err := io.Copy(f, io.LimitReader(resp.Body, 200<<20))
	f.Close()
	head := make([]byte, 2)
	if hf, e := os.Open(setup); e == nil {
		hf.Read(head)
		hf.Close()
	}
	if err != nil || (size > 0 && n != size) || string(head) != "MZ" {
		os.Remove(setup)
		fail(502, "The download didn't finish properly. Try again.")
		return
	}
	// Run the installer quietly. It closes this helper, installs, and starts the new version,
	// which takes over the board window that's already open.
	args := []string{"/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/SP-", "/update=1",
		fmt.Sprintf("/browserpid=%d", browserPID)}
	if err := exec.Command(setup, args...).Start(); err != nil {
		fail(500, "Couldn't start the installer: "+err.Error())
		return
	}
	sendJ(w, 200, map[string]any{"ok": true, "version": rel.Tag})
	if fl, ok := w.(http.Flusher); ok {
		fl.Flush()
	}
	go func() {
		time.Sleep(700 * time.Millisecond)
		if quitFn != nil {
			quitFn() // let go of the program file so the installer can replace it
		}
	}()
}

// relaunch starts the (new) program with the same options and tells it which window to keep watching.
func relaunch(stop func()) func(extra ...string) error {
	return func(extra ...string) error {
		exe := exePath
		var args []string
		for i := 1; i < len(os.Args); i++ {
			a := os.Args[i]
			if strings.HasPrefix(a, "--browser-pid") || strings.HasPrefix(a, "-browser-pid") {
				if !strings.Contains(a, "=") {
					i++
				}
				continue
			}
			args = append(args, a)
		}
		args = append(args, fmt.Sprintf("--browser-pid=%d", browserPID))
		args = append(args, extra...)
		// The new version waits for the port, so start it first: if it won't start, nothing is lost.
		if err := exec.Command(exe, args...).Start(); err != nil {
			return err
		}
		stop()
		os.Exit(0)
		return nil
	}
}

// cleanupOld removes the previous version left behind by an update (it may take a moment to exit).
func cleanupOld() {
	if e, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(e); err == nil {
			e = r
		}
		exePath = e
	}
	exe := exePath
	if exe == "" {
		return
	}
	go func() {
		for i := 0; i < 20; i++ {
			if err := os.Remove(exe + ".old"); err == nil || os.IsNotExist(err) {
				return
			}
			time.Sleep(time.Second)
		}
	}()
}
