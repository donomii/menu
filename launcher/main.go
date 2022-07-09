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
var preserveWindow = true
var glfwInitialized = false
var quitAfter = false

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
		ActiveBufferInsert(ed, "\n?> ")
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
	case "HideWindow":
		ForceHide()

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

				ForceHide()

				return

			}(selected)
			//FIXME some kind of transition here?
			mode = "searching"
			input = ""
			status = ""
			update = true
			selected = 0
		} else {
			toggleWindow()
		}

	}

}
func handleKeys(window *glfw.Window) {
	EscapeKeyCode := 256
	MacEscapeKeyCode := 53

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		/*if key == 301 {
			hideWindow()
			return
		}
		*/
		if action > 0 {

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

//Pushes an existing window to the front.  Window must exist.
func popWindow() {
	log.Println("Popping window")
	update = true
	window.Restore()
	window.Show()
	window.Focus()
}

//Hides an existing window.  Window must exist.
func hideWindow() {
	log.Println("Hiding window")
	window.Iconify()
	window.Hide()
}

//If the window exists, pop it to the front.  Otherwise create it, then pop it.
func ForceFront() {
	if preserveWindow {
		popWindow()
	} else {
		update = true
		createWin = true
	}
}

//If the window exists, hide it.  Otherwise create it, then hide it.
//If preserveWindow is not true, the program will quit.
func ForceHide() {
	if preserveWindow {
		log.Println("Hiding window")
		hideWindow()
	} else {
		log.Println("Exiting")
		update = true
		createWin = false
	}
}
func toggleWindow() {
	log.Println("Toggling window")
	wantWindow = !wantWindow
	if wantWindow {

		if preserveWindow {
			popWindow()
		} else {
			update = true
			createWin = true
		}
	} else {
		if preserveWindow {
			hideWindow()
		} else {
			//This exits the program.
			createWin = false
		}
	}
}
func main() {
	if runtime.GOOS == "darwin" {
		preserveWindow = false
	}
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.BoolVar(&quitAfter, "quit-after", false, "Quit after command or focus loss")
	flag.Parse()

	go WatchKeys()

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
	edWidth = mode.Width - int(float64(mode.Width)*0.2)
	edHeight = mode.Height / 3.0

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
	window.SetPos(mode.Width/10.0, mode.Height/4.0)
	popWindow()
	log.Println("Make glfw window context current")
	window.MakeContextCurrent()
	log.Println("Allocate memory")
	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)
	log.Println("Set up key handlers")
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
		if glfwInitialized {
			hasFocus := glfw.GetCurrentContext().GetAttrib(glfw.Focused)

			if hasFocus == 0 {
				log.Println("Window lost focus")
				createWin = false
			}
		}
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
