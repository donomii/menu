package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	//"time"

	"github.com/donomii/goof"
)

func main() {
	homeDir := goof.HomeDirectory()
	pidfile := homeDir + "/" + "universalmenu.pid"
	launchDir := goof.ExecutablePath()
	if goof.Exists(pidfile) {
		pidString, _ := ioutil.ReadFile(pidfile)
		pid, _ := strconv.Atoi(string(pidString))
		proc, _ := os.FindProcess(pid)
		proc.Kill()
		os.Remove(pidfile)
	} else {
		cmd := exec.Command(launchDir + "/universal_menu_main")
		cmd.Start()
		//time.Sleep((10 * time.Second))
	}
}
