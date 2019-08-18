package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	//"time"

	"github.com/donomii/goof"
)

func pidPath() string {
	homeDir := goof.HomeDirectory()
	pidfile := homeDir + "/" + "universalmenu.pid"
	return pidfile
}

func main() {
	pidfile := pidPath()
	launchDir := goof.ExecutablePath()
	if goof.Exists(pidfile) {
		pidString, _ := ioutil.ReadFile(pidfile)
		pid, _ := strconv.Atoi(string(pidString))
		proc, _ := os.FindProcess(pid)
		proc.Kill()
		os.Remove(pidfile)
	} else {
		cmd := exec.Command(launchDir + "/universal_menu_main.exe")
		cmd.Start()
	}
}
