// gui.go
package main

import (
	"github.com/donomii/glim"

	"fmt"

	"log"
)

var foreColour, backColour *glim.RGBA

func renderEd(w, h int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in renderEd", r)
		}
	}()

	left := 0
	top := 0

	if ed != nil {
		log.Println("Starting editor draw")

		size := w * h * 4
		log.Println("Clearing", size, "bytes(", w, "x", h, ")")
		backColour = &glim.RGBA{0, 0, 0, 255}
		foreColour = &glim.RGBA{255, 255, 255, 255}
		for i := 0; i < size; i = i + 4 {
			pic[i] = ((*backColour)[0])
			pic[i+1] = ((*backColour)[1])
			pic[i+2] = ((*backColour)[2])
			pic[i+3] = ((*backColour)[3])
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
