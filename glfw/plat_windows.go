// +build windows

package main

//#include "c/keywatcher.c"
import "C"

func WatchKeys() {
	//wantWindow = false
	C.watchKeys()
}
