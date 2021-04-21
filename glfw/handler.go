// +build !linux

package main

import "C"
import "log"

//export HandleKey
func HandleKey(k C.int) {
	log.Printf("Key id: %v\n", k)
	if k == 123 || k == 111 {
		toggleWindow()

	}
}
