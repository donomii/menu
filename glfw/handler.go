// +build !linux

package main

import "C"
import (
	"log"
	"runtime"
)

//export HandleKey
func HandleKey(k C.int) bool {
	log.Printf("Global key id: %v\n", k)
	if (runtime.GOOS != "darwin" && k == 123) || // F12 on windows
		k == 111 || //F12 on mac
		k == 177 { //Spotlight key on mac

		toggleWindow()
		return true

	}
	return false
}
