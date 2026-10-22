package main

// The helper stays running while a board page is open, and quits on its own once it isn't.
//
// It can't simply wait for the browser process it started: Edge often hands a new window to an
// Edge that's already running and exits straight away. Instead the page checks in every
// 20 seconds and says goodbye (/api/bye) when it closes or reloads.

import (
	"net/http"
	"sync"
	"time"
)

var (
	lifeMu   sync.Mutex
	lastSeen = time.Now()
	byeAt    time.Time
)

const (
	firstVisitWait = 2 * time.Minute // time for the browser to start and load the board
	quietLimit     = 3 * time.Minute // no word from any page for this long: everything is closed
	byeGrace       = 8 * time.Second // after a goodbye, a reload or another window may still check in
)

// track notes every request from a page, and handles the goodbye.
func track(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lifeMu.Lock()
		if r.URL.Path == "/api/bye" {
			byeAt = time.Now()
		} else {
			lastSeen = time.Now()
		}
		lifeMu.Unlock()
		if r.URL.Path == "/api/bye" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// waitUntilClosed blocks until no board page has been open for a while.
func waitUntilClosed() {
	started := time.Now()
	seenAny := false
	for {
		time.Sleep(2 * time.Second)
		lifeMu.Lock()
		last, bye := lastSeen, byeAt
		lifeMu.Unlock()
		if last.After(started) {
			seenAny = true
		}
		now := time.Now()
		switch {
		case !seenAny && now.Sub(started) > firstVisitWait:
			return
		case seenAny && now.Sub(last) > quietLimit:
			return
		case !bye.IsZero() && bye.After(last) && now.Sub(bye) > byeGrace:
			return
		}
	}
}
