package main

import (
	"os"
	"runtime/debug"

	"github.com/donomii/menu"

	"fmt"
	"io/ioutil"
	"runtime"

	"golang.org/x/image/font/gofont/goregular"

	//"unsafe"

	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"

	//"text/scanner"

	"flag"

	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/donomii/glim"
)

var shell string
var active bool
var myMenu Menu

type Menu []string

var form *glim.FormatParams
var lasttime float64
var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var workerChan chan string
var needsRedraw bool
var atlas *nk.FontAtlas
var sansFont *nk.Font

type Option uint8

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

type UserConfig struct {
	Red, Green, Blue int
}

var winWidth = 600
var winHeight = 400

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)
}

var ed *GlobalConfig
var confFile string
var pic []uint8
var picBytes []byte

var pred, cmd []string
var input string
var selected int
var update bool

func seq(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func UpdateBuffer(ed *GlobalConfig, input string) {
	ClearActiveBuffer(ed)

	ActiveBufferInsert(ed, input)
	ActiveBufferInsert(ed, "\n\n")
	pred, cmd = menu.Predict([]byte(input)) //FIXME
	fmt.Printf("predictions %+v\n", pred)
	for _, v := range seq(selected, len(pred)-1) {
		if v == selected {
			ActiveBufferInsert(ed, "\n")
			ActiveBufferInsert(ed, "        "+pred[v]+"\n\n")
		} else {
			ActiveBufferInsert(ed, "        "+pred[v]+"\n")
		}
	}

	for _, v := range seq(0, selected-1) {
		if v == selected {
			ActiveBufferInsert(ed, "\n")
			ActiveBufferInsert(ed, pred[v]+"\n\n")
		} else {
			ActiveBufferInsert(ed, "        "+pred[v]+"\n")
		}
	}
}
func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	pic = make([]uint8, 3000*3000*4)
	picBytes = make([]byte, 3000*3000*4)
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.StringVar(&shell, "shell", "/bin/bash", "The command shell to run")
	flag.Parse()
	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	foreColour = &glim.RGBA{255, 255, 255, 255}
	backColour = &glim.RGBA{0, 0, 0, 255}

	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)

	//Nuklear

	if err := glfw.Init(); err != nil {
		closer.Fatalln(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err := glfw.CreateWindow(winWidth, winHeight, "Menu", nil, nil)
	if err != nil {
		closer.Fatalln(err)
	}
	win.MakeContextCurrent()

	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		fmt.Printf("Got key %c,%v,%v,%v", key, key, mods, action)
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
				menu.Activate(cmd[selected])
				time.Sleep(1 * time.Second)
				os.Exit(0)
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

	win.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {

		text := fmt.Sprintf("%c", char)
		fmt.Printf("Text input: %v\n", text)
		input = input + text
		UpdateBuffer(ed, input)
		update = true

	})

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas = nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont = nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	} else {
		panic("Font load failed")
	}

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	state := &State{
		bgColor: nk.NkRgba(255, 255, 255, 255),
	}

	fpsTicker := time.NewTicker(time.Second / 30)

	SetFont(ed.ActiveBuffer, 12)
	log.Println("Starting main loop")
	needsRedraw = true

	for {
		select {

		case <-exitC:
			fmt.Println("Shutdown")
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case <-fpsTicker.C:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			winWidth, winHeight = win.GetSize()
			if needsRedraw {
				lasttime = glfw.GetTime()

				gfxMain(win, ctx, state)
				needsRedraw = false

				active = true

			} else {
				TARGET_FPS := 10.0
				if glfw.GetTime() < lasttime+1.0/TARGET_FPS {
					time.Sleep(10 * time.Millisecond)
					//runtime.GC()
				} else {
					needsRedraw = true
				}
			}

		}

	}

}
