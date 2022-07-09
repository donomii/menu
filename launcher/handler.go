//go:build !linux
// +build !linux

package main

import "C"
import (
	"log"
	"runtime"
)

//export HandleKey
func HandleKey(key C.int) bool {
	log.Printf("Global key id: %v\n", key)
	update = true

	if (runtime.GOOS != "darwin" && key == 123) || // F12 on windows
		key == 111 || //F12 on mac
		key == 177 { //Spotlight key on mac

		if createWin {
			ForceHide()
		} else {
			ForceFront()
		}
		return true

	}

	if key == 53 {
		//os.Exit(0)
		log.Println("Escape pressed")
		doKeyPress("HideWindow")
		return true
	}
	if key == 265 {
		doKeyPress("SelectPrevious")
		return true
	}

	if key == 264 {
		doKeyPress("SelectNext")
		return true
	}

	if key == 257 {
		doKeyPress("Activate")
		return true
	}

	if key == 259 {
		doKeyPress("Backspace")
		return true

	}

	return false
}
