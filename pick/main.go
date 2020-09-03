package main

import (
	"flag"
	"fmt"

	"runtime"
	"time"

	"github.com/donomii/glim"

	"io/ioutil"
	"log"
	"os"

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

func Seq(min, max int) []int {
	size := max - min + 1
	if size < 1 {
		return []int{}
	}
	a := make([]int, size)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func UpdateBuffer(ed *GlobalConfig, input string) {
	searchString = input
	results = search(input, 10)
	ClearActiveBuffer(ed)

	ActiveBufferInsert(ed, "\n ")
	ActiveBufferInsert(ed, input)
	ActiveBufferInsert(ed, "\n\n")
	refreshTerm()
}

func addStr(s string) {
	ActiveBufferInsert(ed, s)
}
func refreshTerm() {

	//statuses["Screen"] = "Refresh"
	ClearActiveBuffer(ed)
	addStr(fmt.Sprintf(" Search for:%v", searchString))

	//prevRecord := ResultRecordTransmittable{"", -1, makeFingerprintFromData(""), "", 0}
	prevRecord := ResultRecordTransmittable{}

	if len(results) > 0 {
		addStr(fmt.Sprintf("     %v (line %v)\n", results[0].Filename, results[0].Line))
		prevRecord = results[0]
	}
	itempos := 0
	for i, elem := range results {

		if !(i == 0) && !(elem.Filename == prevRecord.Filename) {
			addStr(fmt.Sprintf("   %v\n", elem.Filename))

		}
		if itempos == selected {
			addStr("*")
		} else {
			addStr(" ")
		}
		//if elem.Line != "-1" && strings.HasPrefix(elem.Filename, "http") {
		addStr(fmt.Sprintf(" %v", elem.Score))
		addStr(fmt.Sprintf("        (line %v) %v\n", elem.Line, elem.Sample))

		itempos++
		prevRecord = elem
		//}

	}
	addStr(fmt.Sprintf("\n %v results", len(results)))
	addStr(fmt.Sprintf("%v", statuses))
	addStr(fmt.Sprintf("\nType your search terms, add a - to the end of word to remove that word (word-)"))
	addStr(fmt.Sprintf("\nUp/Down Arrows to select a result, Right Arrow to edit that file, Escape Quits"))

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
				mode = "loading"
				results = search(searchStr, 10)
				update = true

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
		results = search(searchStr, 10)
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

	log.Println("Init glfw")
	if err := glfw.Init(); err != nil {
		panic("failed to initialize glfw: " + err.Error())
	}
	defer glfw.Terminate()

	log.Println("Setup window")
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()
	edWidth = mode.Width - int(float64(mode.Width)*0.1)
	//edHeight = mode.Height

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	log.Println("Make window")
	window, err := glfw.CreateWindow(edWidth, edHeight, "Menu", nil, nil)
	if err != nil {
		panic(err)
	}
	log.Println("Make context current")
	window.MakeContextCurrent()
	log.Println("Allocate memory")
	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)
	log.Println("Set up handlers")
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

	log.Println("Init gl")
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
	go startPick()
	log.Println("Start rendering")
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

	gl.Viewport(0, 0, int32(w)*screenScale(), int32(h)*screenScale())
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
