package main

// Aircraft data for the board: from the user's own ADS-B receiver, or, for people without one,
// from a free community ADS-B network around the radar center.

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Online services. All answer in the readsb format ({"ac":[...]}) that tar1090 also uses.
// Each asks for no more than about one request per second; the helper asks at most every 5 s.
var onlineADSB = map[string]string{
	"adsb.lol":       "https://api.adsb.lol/v2/point/%.3f/%.3f/%d",
	"airplanes.live": "https://api.airplanes.live/v2/point/%.3f/%.3f/%d",
	"adsb.fi":        "https://opendata.adsb.fi/api/v3/lat/%.3f/lon/%.3f/dist/%d",
}

const defaultOnline = "adsb.lol"

// adsbSource is "online" or "receiver". Boards set up with a receiver address before the
// online option existed keep using their receiver.
func adsbSource(s map[string]any) string {
	if v, _ := s["adsbSource"].(string); v == "online" || v == "receiver" {
		return v
	}
	if u, _ := s["adsbUrl"].(string); strings.TrimSpace(u) != "" {
		return "receiver"
	}
	return "online"
}

func num(s map[string]any, path ...string) (float64, bool) {
	var cur any = s
	for _, k := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return 0, false
		}
		cur = m[k]
	}
	f, ok := cur.(float64)
	return f, ok
}

func handleAdsb(w http.ResponseWriter, r *http.Request) {
	s := readSettings()
	if adsbSource(s) == "receiver" {
		u, _ := s["adsbUrl"].(string)
		u = strings.TrimSpace(u)
		if u == "" {
			sendS(w, 400, "No receiver address set. Add it in Settings > Flights, or choose Online.", "text/plain")
			return
		}
		st, body, _ := fetch(u, 1500*time.Millisecond, 4*time.Second)
		ct := "text/plain"
		if st == 200 {
			ct = "application/json"
		}
		send(w, st, body, ct)
		return
	}

	// Online: center on the radar center the page sends, else the saved one.
	q := r.URL.Query()
	lat, e1 := strconv.ParseFloat(q.Get("lat"), 64)
	lon, e2 := strconv.ParseFloat(q.Get("lon"), 64)
	if e1 != nil || e2 != nil {
		var ok1, ok2 bool
		lat, ok1 = num(s, "home", "lat")
		lon, ok2 = num(s, "home", "lon")
		if !ok1 || !ok2 {
			sendS(w, 400, "Set your location in Settings > Location first.", "text/plain")
			return
		}
	}
	if math.IsNaN(lat) || math.IsNaN(lon) || math.Abs(lat) > 90 || math.Abs(lon) > 180 {
		sendS(w, 400, "Bad location.", "text/plain")
		return
	}
	radius, err := strconv.Atoi(q.Get("r"))
	if err != nil {
		radius = 40
	}
	radius = max(10, min(250, radius))

	prov, _ := s["adsbOnline"].(string)
	pattern, ok := onlineADSB[prov]
	if !ok {
		prov, pattern = defaultOnline, onlineADSB[defaultOnline]
	}
	u := fmt.Sprintf(pattern, lat, lon, radius)
	st, body, _ := fetch(u, 5*time.Second, 10*time.Second)
	if st != 200 {
		sendS(w, st, prov+": "+string(body), "text/plain")
		return
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(body, &raw) != nil {
		sendS(w, 502, prov+" sent something that isn't aircraft data.", "text/plain")
		return
	}
	list := raw["ac"]
	if len(list) == 0 {
		list = raw["aircraft"]
	}
	if len(list) == 0 || string(list) == "null" {
		list = json.RawMessage("[]")
	}
	out, _ := json.Marshal(map[string]any{
		"now":      float64(time.Now().UnixMilli()) / 1000,
		"aircraft": list,
		"source":   prov,
	})
	send(w, 200, out, "application/json")
}
