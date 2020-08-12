package main

import (
	"flag"
	"fmt"

	"runtime"
	"time"

	"github.com/donomii/glim"
	"github.com/donomii/goof"

	"io/ioutil"
	"log"
	"os"

	"github.com/donomii/menu"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

func init() { runtime.LockOSThread() }

var ed *GlobalConfig
var confFile string
var pic []uint8

var pred []string
var input, status string
var selected int
var update bool = true
var form *glim.FormatParams
var edWidth = 1000
var edHeight = 500
var mode = "searching"

func UpdateBuffer(ed *GlobalConfig, input string) {
	ClearActiveBuffer(ed)

	if mode == "searching" {
		ActiveBufferInsert(ed, "\n\n")
		ActiveBufferInsert(ed, input)
		ActiveBufferInsert(ed, "\n\n")
		pred = menu.Predict([]byte(input))
		log.Printf("predictions %+v\n", pred)
		if len(pred) > 0 {
			for _, v := range goof.Seq(selected, len(pred)-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n")
					ActiveBufferInsert(ed, "        "+pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}

			for _, v := range goof.Seq(0, selected-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n")
					ActiveBufferInsert(ed, pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}
		}
	} else {
		ActiveBufferInsert(ed, "?\n\n")
		ActiveBufferInsert(ed, status)
	}
}

func handleKeys(window *glfw.Window) {
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		//fmt.Printf("Got key %c,%v,%v,%v", key, key, mods, action)
		if action > 0 {
			if key == 256 {
				os.Exit(0)
			}

			if key == 265 {
				selected -= 1
				if selected < 0 {
					selected = 0
				}
			}

			if key == 264 {
				selected += 1
				if selected > len(pred)-1 {
					selected = len(pred) - 1
				}
			}

			if key == 257 {

				status = "Loading " + pred[selected]
				mode = "laoding"
				update = true
				go func() {
					menu.Activate(pred[selected])
					time.Sleep(1 * time.Second)
					os.Exit(0)
				}()
			}

			if key == 259 {
				if len(input) > 0 {
					input = input[0 : len(input)-1]
				}
			}

			UpdateBuffer(ed, input)
			update = true
		}

	})

	window.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {

		text := fmt.Sprintf("%c", char)
		//fmt.Printf("Text input: %v\n", text)
		input = input + text
		UpdateBuffer(ed, input)
		update = true

	})
}

func main() {
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.Parse()

	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if err := glfw.Init(); err != nil {
		panic("failed to initialize glfw: " + err.Error())
	}
	defer glfw.Terminate()

	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(edWidth, edHeight, "Menu", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	handleKeys(window)

	//This should be SetFramebufferSizeCallback, but that doesn't work, so...
	window.SetSizeCallback(func(w *glfw.Window, width int, height int) {

		edWidth = width
		edHeight = height
		renderEd(edWidth, edHeight)
		blit(pic, edWidth, edHeight)
		window.SwapBuffers()
		update = true
	})

	if err := gl.Init(); err != nil {
		panic(err)
	}
	/*
		go func() {
			lastTime := glfw.GetTime()

			for {
				nowTime := glfw.GetTime()
				if nowTime-lastTime < 10000.0 {

					update = true
					fmt.Println("Forece refresh")
				} else {
					return
				}
			}
		}()
	*/

	lastTime := glfw.GetTime()
	frames := 0
	UpdateBuffer(ed, input)
	for !window.ShouldClose() {
		time.Sleep(35 * time.Millisecond)
		frames++
		nowTime := glfw.GetTime()
		if nowTime-lastTime >= 1.0 {
			//status = fmt.Sprintf("%.3f ms/f  %.0ffps\n", 1000.0/float32(frames), float32(frames))
			frames = 0
			lastTime += 1.0
		}

		if update {
			renderEd(edWidth, edHeight)
			blit(pic, edWidth, edHeight)
			window.SwapBuffers()
			update = false
		}
		glfw.PollEvents()
	}
}

func blit(pix []uint8, w, h int) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Viewport(0, 0, int32(w)*2, int32(h)*2)
	gl.Ortho(0, 1, 1, 0, 0, -1)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D, 0,
		gl.RGBA,
		int32(w), int32(h), 0,
		gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(pix),
	)

	gl.Enable(gl.TEXTURE_2D)
	{
		gl.Begin(gl.QUADS)
		{
			gl.TexCoord2i(0, 0)
			gl.Vertex2i(0, 0)

			gl.TexCoord2i(1, 0)
			gl.Vertex2i(1, 0)

			gl.TexCoord2i(1, 1)
			gl.Vertex2i(1, 1)

			gl.TexCoord2i(0, 1)
			gl.Vertex2i(0, 1)
		}
		gl.End()
	}
	gl.Disable(gl.TEXTURE_2D)

	gl.Flush()

	gl.DeleteTextures(1, &texture)
}
