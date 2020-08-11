package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/BurntSushi/toml"
	"github.com/donomii/menu"

	"golang.org/x/image/font/gofont/goregular"

	//"unsafe"

	"sync"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"

	//"text/scanner"

	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/donomii/glim"
	"github.com/donomii/goof"
	"github.com/rivo/tview"
)

var AppMode int
var form *glim.FormatParams
var demoText = "hi"
var displaySplit string = "None"
var result = ""
var EditStr = `lalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalala`
var EditBytes []byte
var userbytes []byte
var lastUserbytes []byte
var optionsList []string
var tokens [][]string
var atlas *nk.FontAtlas

var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var lastElemSelected string
var lastElemSelectedIndex int
var app *tview.Application
var workerChan chan string
var currentNodeLock sync.Mutex
var fontSmall *nk.Font
var fontLarge *nk.Font

var activeSelection = 0

var currentNode *menu.Node

func getCurrentNode() *menu.Node {
	currentNodeLock.Lock()
	n := currentNode
	currentNodeLock.Unlock()
	return n
}

var currentThing []*menu.Node

type Menu []string

func UberMenu() *menu.Node {
	node := menu.MakeNodeLong("Main menu",
		[]*menu.Node{
			menu.AppsMenu(),
			menu.HistoryMenu(),
			//gitMenu(),
			//gitHistoryMenu(),
			//fileManagerMenu(),
			//controlMenu(),
		},
		"", "")
	return node
}

var myMenu Menu

func configFile() *menu.Node {
	return menu.MakeNodeShort("Edit Config", []*menu.Node{})
}

func gitHistoryMenu() *menu.Node {
	node := menu.MakeNodeShort("previous git commands", []*menu.Node{})
	menu.AddTextNodesFromString(node, goof.Grep("git", goof.QC([]string{"fish", "-c", "history"})))
	return node
}

var header string

type Form struct {
	children []*Form
	val      string
}

type Option uint8

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

type UserConfig struct {
	Red, Green, Blue int
}

var winWidth = 900
var winHeight = 900
var ed *menu.GlobalConfig
var config UserConfig
var confFile string

var lastKey time.Time

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)
	lastKey = time.Now()
	//log.Println("Locked to main thread")
}

func pidPath() string {
	homeDir := goof.HomeDirectory()
	pidfile := homeDir + "/" + "universalmenu.pid"
	return pidfile
}

func togglePidFile() {
	if goof.Exists(pidPath()) {
		fmt.Println("Found lockfile at", pidPath(), ", exiting")
		os.Exit(1)
	} else {
		pidStr := fmt.Sprintf("%v", os.Getpid())
		log.Printf("Writing pid to %v\n", pidStr)
		ioutil.WriteFile(pidPath(), []byte(pidStr), 0644)
	}
}

func main() {
	userbytes = []byte("                                                                                          ")
	//	runtime.LockOSThread()
	runtime.GOMAXPROCS(4)

	confFile = goof.ConfigFilePath(".menu.json")
	log.Println("Loading config from:", confFile)
	configBytes, conferr := ioutil.ReadFile(confFile)
	if conferr != nil {
		log.Println("Writing fresh config to:", confFile)
		ioutil.WriteFile(confFile, []byte("test"), 0644)
		configBytes, conferr = ioutil.ReadFile(confFile)
	}
	var force bool
	toml.Decode(string(configBytes), &config)
	flag.BoolVar(&force, "force", false, "Ignore lockfile and start")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()
	if !force {
		togglePidFile()
	}
	go func() {
		ed = menu.NewEditor()
		//Create a text formatter
		form = glim.NewFormatter()

	}()

	//currentNode = addTextNodesFromStrStrStr(currentNode, MailSummaries())

	//currentNode =
	currentThing = []*menu.Node{getCurrentNode()}
	//result := ""

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

	width, height := win.GetSize()
	//log.Printf("glfw: created window %dx%d", width, height)

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	gl.Viewport(0, 0, int32(width-1), int32(height-1))

	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		if mods == 2 && action == 1 && key != 341 {
			mask := ^byte(64 + 128)
			log.Printf("key mask: %#b", mask)
			val := byte(key)
			log.Printf("key val: %#b", val)
			b := mask & val
			log.Printf("key byte: %#b", b)

		}

		if action == 0 && mods == 0 {
			switch key {
			case 257: //Enter
				if activate(activeSelection, comboCallback(userbytes, lastUserbytes)[activeSelection]) {
					os.Remove(pidPath())
					os.Exit(0)
				}
			case 256: //Escape
				os.Remove(pidPath())
				os.Exit(0)
			}
		}
	})

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas = nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	/*data, err := ioutil.ReadFile("FreeSans.ttf")
	if err != nil {
		panic("Could not find file")
	}*/

	fontSmall = nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 16, nil)
	fontLarge = nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 32, nil)
	// sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()

	nk.NkStyleSetFont(ctx, fontSmall.Handle())

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	state := &State{
		bgColor: nk.NkRgba(28, 48, 62, 255),
	}
	fpsTicker := time.NewTicker(time.Second / 30)
	//
	for {
		select {
		case <-exitC:
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
			//log.Printf("glfw: created window %dx%d", width, height)
			gfxMain(win, ctx, state)
			runtime.GC()
		}
	}

	//End Nuklear

	if ui {
		for {

			//currentNode, currentThing, result = doui(currentNode, currentThing, result)
		}
	}
}

func b(v int32) bool {
	return v == 1
}

func fflag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
