// +build windows

package main

import "C"
import "fmt"

//export HandleKey
func HandleKey(k C.int) {
	fmt.Printf("Key id: %v\n", k)
	if k == 161 {
		popWindow()
	}
}
