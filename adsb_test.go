package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTest(t *testing.T, settings string) {
	dir := t.TempDir()
	dataDir, settingsPath, picturesDir = dir, filepath.Join(dir, "settings.json"), filepath.Join(dir, "pictures")
	os.WriteFile(settingsPath, []byte(settings), 0o644)
	cache = map[string]cached{}
}

func TestOnlineADSB(t *testing.T) {
	var gotPath string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"ac":[{"hex":"a6f486","flight":"ASA123 ","lat":47.95,"lon":-122.3,"alt_baro":3200,"seen":1}],"msg":"No error","now":1761100000000,"total":1}`))
	}))
	defer fake.Close()
	old := onlineADSB["adsb.lol"]
	onlineADSB["adsb.lol"] = fake.URL + "/v2/point/%.3f/%.3f/%d"
	defer func() { onlineADSB["adsb.lol"] = old }()

	setupTest(t, `{}`) // brand-new install: online by default
	rec := httptest.NewRecorder()
	handleAdsb(rec, httptest.NewRequest("GET", "/api/adsb?lat=47.9445&lon=-122.3046&r=38", nil))
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if gotPath != "/v2/point/47.944/-122.305/38" && gotPath != "/v2/point/47.945/-122.305/38" {
		t.Fatalf("asked the wrong place: %s", gotPath)
	}
	var out struct {
		Aircraft []map[string]any `json:"aircraft"`
		Source   string           `json:"source"`
	}
	json.Unmarshal(rec.Body.Bytes(), &out)
	if len(out.Aircraft) != 1 || out.Aircraft[0]["hex"] != "a6f486" || out.Source != "adsb.lol" {
		t.Fatalf("bad answer: %s", rec.Body)
	}
}

func TestSourceChoice(t *testing.T) {
	cases := map[string]string{
		`{}`: "online",
		`{"adsbUrl":"http://piaware.local/skyaware/data/aircraft.json"}`: "receiver",
		`{"adsbUrl":"http://x/aircraft.json","adsbSource":"online"}`:     "online",
		`{"adsbSource":"receiver"}`:                                      "receiver",
	}
	for in, want := range cases {
		var m map[string]any
		json.Unmarshal([]byte(in), &m)
		if got := adsbSource(m); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}

func TestReceiverStillWorks(t *testing.T) {
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"now":1,"aircraft":[{"hex":"abc123"}]}`))
	}))
	defer fake.Close()
	setupTest(t, `{"adsbUrl":"`+fake.URL+`/data/aircraft.json"}`)
	rec := httptest.NewRecorder()
	handleAdsb(rec, httptest.NewRequest("GET", "/api/adsb?lat=1&lon=2&r=30", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "abc123") {
		t.Fatalf("receiver: %d %s", rec.Code, rec.Body)
	}
}
