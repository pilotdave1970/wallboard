//go:build !windows

package main

// Start with Windows only exists on Windows; these keep the code building elsewhere for tests.
var fakeStartup bool

func startupEnabled() bool     { return fakeStartup }
func setStartup(on bool) error { fakeStartup = on; return nil }
