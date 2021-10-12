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
	//wantWindow = false
	log.Println("Starting keywatcher")
	C.watchKeys()
}

func screenScale() int32 { return 2 }
