package main

import (
	"os/exec"
	"syscall"
)

const runKey = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
const runName = "JetsWallboard"

func hiddenCmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	return c
}

func startupEnabled() bool {
	return hiddenCmd("reg.exe", "query", runKey, "/v", runName).Run() == nil
}

// setStartup adds or removes the wallboard from the programs Windows starts at sign-in.
// Full screen or not is decided by the board's own setting when it opens.
func setStartup(on bool) error {
	if !on {
		if !startupEnabled() {
			return nil
		}
		return hiddenCmd("reg.exe", "delete", runKey, "/v", runName, "/f").Run()
	}
	return hiddenCmd("reg.exe", "add", runKey, "/v", runName, "/t", "REG_SZ", "/d", `"`+exePath+`"`, "/f").Run()
}
