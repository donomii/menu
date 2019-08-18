package main

import (
	"io/ioutil"
	"log"
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
			log.Println("Removing pid file to prevent locking user out of menu")
			os.Remove(pidfile)
		}
	}()
	launchDir := goof.ExecutablePath()
	if goof.Exists(pidfile) {
		pidString, _ := ioutil.ReadFile(pidfile)
		pid, _ := strconv.Atoi(string(pidString))
		proc, err := os.FindProcess(pid)
		if err == nil {
			proc.Kill()
		}
		os.Remove(pidfile)
	} else {
		cmd := exec.Command(launchDir + "/universal_menu_main.exe")
		cmd.Start()
	}
}
