// gui.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"

	"github.com/donomii/glim"
)

var foreColour, backColour *glim.RGBA

func renderEd(w, h int) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			fmt.Println("Recovered in renderEd", r)

		}
	}()
	left := 0
	top := 0

	if ed != nil {
		log.Println("Starting editor draw")

		size := w * h * 4
		log.Println("Clearing", size, "bytes(", w, "x", h, ")")
		backColour = &glim.RGBA{0, 0, 0, 0}
		patternColour := &glim.RGBA{128, 100, 100, 255}
		foreColour = &glim.RGBA{255, 255, 255, 255}
		num := rand.Intn(100)
		log.Println("Background rand:", num)
		for i := 0; i < size; i = i + 4 {
			var Colour *glim.RGBA
			if (i^int(i/w))%3 == 0 {
				//if !(int(math.Pow(float64(i), float64(int(i/w))))%9>0){
				Colour = patternColour
			} else {
				Colour = backColour
			}
			pic[i] = ((*Colour)[0])
			pic[i+1] = ((*Colour)[1])
			pic[i+2] = ((*Colour)[2])
			pic[i+3] = ((*Colour)[3])
		}

		form = ed.ActiveBuffer.Formatter
		form.Colour = foreColour
		form.Outline = true

		mouseX := 10
		mouseY := 10
		displayText := ed.ActiveBuffer.Data.Text

		log.Println("Render paragraph", string(displayText))

		ed.ActiveBuffer.Formatter.FontSize = 32
		glim.RenderPara(ed.ActiveBuffer.Formatter,
			0, 0, 0, 0,
			w, h, w, h,
			int(mouseX)-left, int(mouseY)-top, pic, displayText,
			false, true, false)
		log.Println("Finished render paragraph")
	}
}
