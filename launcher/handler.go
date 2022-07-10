//go:build !linux
// +build !linux

package main

import "C"
import (
	"fmt"
	"log"
	"runtime"
)

//export HandleKey
func HandleKey(key C.int, down C.int) C.uchar {

	log.Printf("Global key id: %v\n", key)
	update = true

	if (runtime.GOOS != "darwin" && key == 123) || // F12 on windows
		key == 111 || //F12 on mac
		key == 177 || //Spotlight key on mac
		key == 109 { //F10

		if down != 1 {
			fmt.Println("  Golang returning true in F12")
			return C.uchar(1)
		}

		if wantWindow {
			ForceHide()
		} else {
			ForceFront()
		}
		fmt.Println("  Golang returning true in F12")
		return C.uchar(1)

	}

	if wantWindow {
		if key == 53 {
			if down != 1 {
				fmt.Println("  Golang returning true in Esc")
				return C.uchar(1)
			}
			//os.Exit(0)
			log.Println("Escape pressed")
			doKeyPress("HideWindow")
			fmt.Println("  Golang returning true in Esc")
			return C.uchar(1)
		}
	}

	fmt.Println("  Golang returning false")
	return C.uchar(0)
}
