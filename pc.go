package main

// "This PC" settings, changed from Settings > Board on the wallboard itself:
//   start with Windows, open full screen, and let phones/tablets at home open the board.

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func settingBool(key string) bool {
	b, _ := readSettings()[key].(bool)
	return b
}

func handlePC(w http.ResponseWriter, r *http.Request, lanNow bool) {
	if r.Method == "POST" {
		var req struct {
			Startup    *bool `json:"startup"`
			Fullscreen *bool `json:"fullscreen"`
			Lan        *bool `json:"lan"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil {
			sendS(w, 400, "Bad request.", "text/plain")
			return
		}
		if req.Startup != nil {
			if err := setStartup(*req.Startup); err != nil {
				sendS(w, 500, "Couldn't change Windows startup: "+err.Error(), "text/plain")
				return
			}
		}
		lock.Lock()
		s := readSettings()
		if req.Fullscreen != nil {
			s["pcFullscreen"] = *req.Fullscreen
		}
		if req.Lan != nil {
			s["pcLan"] = *req.Lan
		}
		err := writeSettings(s)
		lock.Unlock()
		if err != nil {
			sendS(w, 500, "Couldn't save: "+err.Error(), "text/plain")
			return
		}
		restart := req.Lan != nil && *req.Lan != lanNow
		sendJ(w, 200, map[string]any{"ok": true, "restarting": restart})
		if restart && restartFn != nil {
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
			go func() { time.Sleep(500 * time.Millisecond); restartFn() }()
		}
		return
	}
	sendJ(w, 200, map[string]any{
		"startup":    startupEnabled(),
		"fullscreen": settingBool("pcFullscreen"),
		"lan":        lanNow,
	})
}
