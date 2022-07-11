package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"runtime"

	"time"

	menu ".."
	"github.com/donomii/glim"
	"github.com/donomii/goof"

	"io/ioutil"
	"log"

	//".."
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() { runtime.LockOSThread() }

var ed *GlobalConfig
var confFile string
var pic []uint8

var pred []string
var predAction []string
var input, status string
var selected int
var update bool = true
var form *glim.FormatParams
var edWidth = 700
var edHeight = 500
var mode = "searching"
var window *glfw.Window

var wantWindow = true
var createWin = true
var glfwInitialized = false
var quitAfter = true
var screenScale = int32(2)
var mouseX, mouseY, lastMouseX, lastMouseY, dragOffSetX, dragOffSetY, globalMouseX, globalMouseY float64
var dragStartX, dragStartY float64
var windowPosX, windowPosY int
var mouseDrag = false

var origPosX, origPosY int

var title = "Block"
var command = []string{"osascript", "-e", "tell app \"Terminal\" to do script \"echo hello\""}
var fontSize = float64(16)
var captureOutput = false

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
	ClearActiveBuffer(ed)
	if selected > len(pred)-1 {
		selected = 0
	}
	if mode == "searching" {
		ActiveBufferInsert(ed, "\n"+title)
		ActiveBufferInsert(ed, input)
		ActiveBufferInsert(ed, "\n\n")
		if len(input) > 0 {
			pred, predAction = menu.Predict([]byte(input))

			log.Printf("predictions %#v, %#v\n", pred, predAction)

			pred = append(pred, "Menu Settings")
			predAction = append(predAction, "Menu Settings") //FIXME make this a file:// url
			for _, v := range Seq(selected, len(pred)-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n\n")
					ActiveBufferInsert(ed, "        "+pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}

			for _, v := range Seq(0, selected-1) {
				if v == selected {
					ActiveBufferInsert(ed, "\n")
					ActiveBufferInsert(ed, pred[v]+"\n\n")
				} else {
					ActiveBufferInsert(ed, "        "+pred[v]+"\n")
				}
			}

		}
	} else {
		ActiveBufferInsert(ed, "Loading\n\n")
		ActiveBufferInsert(ed, status)
	}
}

func doKeyPress(action string) {
	switch action {

	case "SelectPrevious":
		selected -= 1
		if selected < 0 {
			selected = 0
		}

	case "SelectNext":
		selected += 1
		if selected > len(pred)-1 {
			selected = len(pred) - 1
		}

	case "Backspace":
		if len(input) > 0 {
			input = input[0 : len(input)-1]
		}

	case "Activate":
		if len(pred) > 0 {
			status = "Loading " + pred[selected] + predAction[selected]
			mode = "loading"
			update = true
			log.Printf("Activating %v\n", pred[selected])
			go func(thread_selected int) {
				value := predAction[thread_selected]
				if strings.HasPrefix(value, "internal://") {
					cmd := strings.TrimPrefix(value, "internal://")
					if cmd == "EditRecallFile" {
						recallFile := menu.RecallFilePath()

						log.Println("Opening for edit: ", recallFile)

						//goof.QC([]string{"open", recallFile})
						go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", recallFile})
						go goof.Command("/usr/bin/open", []string{recallFile})
					}
				}

				menu.Activate(value)

				return

			}(selected)
			//FIXME some kind of transition here?
			mode = "searching"
			input = ""
			status = ""
			update = true
			selected = 0
		}

	}

}
func handleKeys(window *glfw.Window) {
	EscapeKeyCode := 256
	MacEscapeKeyCode := 53
	MacF12KeyCode := 301

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		/*if key == 301 {
			hideWindow()
			return
		}
		*/
		if action > 0 {
			if key == glfw.Key(MacF12KeyCode) || key == 109 {
				doKeyPress("HideWindow")
			}
			//ESC
			if key == glfw.Key(EscapeKeyCode) || key == glfw.Key(MacEscapeKeyCode) {
				//os.Exit(0)
				log.Println("Escape pressed")
				doKeyPress("HideWindow")
				return
			}

			if key == 265 {
				doKeyPress("SelectPrevious")
			}

			if key == 264 {
				doKeyPress("SelectNext")
			}

			if key == 257 {
				doKeyPress("Activate")
			}

			if key == 259 {
				doKeyPress("Backspace")
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

func handleMouse(window *glfw.Window) {

	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft {
			if action == glfw.Press {
				log.Println("Mouse button pressed")
				mouseDrag = true

				dragStartX = globalMouseX
				dragStartY = globalMouseY

				origPosX, origPosY = window.GetPos()
				dragOffSetX = mouseX
				dragOffSetY = mouseY

			}

			if action == glfw.Release {
				deltaX := globalMouseX - dragStartX
				deltaY := globalMouseY - dragStartY
				if goof.AbsFloat64(deltaX) < 10 || goof.AbsFloat64(deltaY) < 10 {
					goof.QC(command)
				}
				mouseDrag = false
			}
		}

		if button == glfw.MouseButtonRight {
			if action == glfw.Release {

			}
		}
	})
}

func handleMouseMove(window *glfw.Window) {

	window.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		//fmt.Printf("Mouse moved to %v,%v\n", xpos, ypos)
		lastMouseX = mouseX
		lastMouseY = mouseY
		mouseX = xpos
		mouseY = ypos
		X, Y := window.GetPos()
		globalMouseX = xpos - float64(X)
		globalMouseY = ypos - float64(Y)

		//log.Printf("Mouse moved to %v,%v\n", mouseX, mouseY)
		if mouseDrag {

			deltaX := globalMouseX - dragStartX
			deltaY := globalMouseY - dragStartY
			log.Printf("deltaX: %v, deltaY: %v, mouseX: %v, mouseY: %v, dragStartX: %v, dragStartY: %v, dragOffSetX: %v, dragOffSetY: %v\n", deltaX, deltaY, mouseX, mouseY, dragStartX, dragStartY, dragOffSetX, dragOffSetY)
			if (deltaX > 10) || (deltaX < -10) || (deltaY > 10) || (deltaY < -10) {
				windowPosX += int(mouseX - dragOffSetX)
				windowPosY += int(mouseY - dragOffSetY)
				log.Printf("Dragged to %v,%v\n", windowPosX, windowPosY)
				window.SetPos(int(windowPosX), int(windowPosY))
			}
		}

	})
}
func main() {

	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.BoolVar(&quitAfter, "quit-after", false, "Quit after command or focus loss")
	flag.BoolVar(&captureOutput, "capture-output", false, "Capture output of command")
	flag.StringVar(&title, "title", title, "Initial text")
	flag.Float64Var(&fontSize, "font-size", fontSize, "Font size")
	flag.Parse()
	c := flag.Args()
	if len(c) == 0 {
		log.Println("No command specified, using default")
		log.Println(`Command example: ./floaty osascript -e "tell app \"Terminal\" to do script \"cd ~/git/menu/floaty && go build .\""`)
	} else {
		command = c
		log.Printf("Command: %v\n", command)
	}

	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	for {

		if createWin {
			createWindow()
		}
		if quitAfter {
			fmt.Println("Quitting after window close")
			os.Exit(0)
		}
		time.Sleep(10 * time.Millisecond)

	}
}

func createWindow() {

	log.Println("Init glfw")
	if err := glfw.Init(); err != nil {
		panic("failed to initialize glfw: " + err.Error())
	}

	defer func() {
		glfwInitialized = false
		glfw.Terminate()
		if quitAfter {
			os.Exit(0)
		}
	}()

	log.Println("Setup window")
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()
	edWidth = 128
	edHeight = 128

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)

	//glfw.WindowHint(glfw.AutoIconify, glfw.True)

	//glfw.WindowHint(glfw.TransparentFramebuffer, 1)
	log.Println("Make window", edWidth, "x", edHeight)
	var err error
	window, err = glfw.CreateWindow(edWidth, edHeight, "Menu", nil, nil)

	if err != nil {
		panic(err)
	}

	log.Println("Set up key handlers")
	handleKeys(window)
	handleMouse(window)
	handleMouseMove(window)

	window.SetPos(mode.Width/10.0, mode.Height/4.0)
	//popWindow()
	log.Println("Make glfw window context current")
	window.MakeContextCurrent()
	log.Println("Allocate memory")
	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, fontSize)

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

	windowPosX, windowPosY = window.GetPos()

	lastTime := glfw.GetTime()
	frames := 0
	UpdateBuffer(ed, input)
	log.Println("Start rendering")
	glfwInitialized = true
	for !window.ShouldClose() && createWin {
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
	log.Println("Normal glfw shutdown")
	if quitAfter {
		fmt.Println("Quitting after window close")
		os.Exit(0)
	}
}

func blit(pix []uint8, w, h int) {
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Viewport(0, 0, int32(w)*screenScale, int32(h)*screenScale)
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
