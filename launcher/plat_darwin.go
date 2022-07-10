//go:build darwin
// +build darwin

package main

import (
	"log"
)

//#include "c/macGlobalKeyHook.c"
// #cgo LDFLAGS: -framework ApplicationServices
// #cgo CFLAGS: -mmacosx-version-min=11.0.0
import "C"

func WatchKeys() {
	if !quitAfter {
		log.Println("Starting keywatcher")
		C.watchKeys()
	} else {
		log.Println("QuitAfter detected, not starting keywatcher")
	}
}

func ReEnableEventTap() {
	log.Println("Re-enabling event tap")
	C.ReEnableEventTap()
}

func DisableEventTap() {
	log.Println("Disabling event tap")
	C.DisableEventTap()
}

func screenScale() int32 { return 2 }
