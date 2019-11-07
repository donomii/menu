package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"

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

var activeSelection = 1

var currentNode *Node

func updateCurrentNode(n *Node) {
	currentNodeLock.Lock()
	currentNode = n
	currentNodeLock.Unlock()
}

func getCurrentNode() *Node {
	currentNodeLock.Lock()
	n := currentNode
	currentNodeLock.Unlock()
	return n
}

var currentThing []*Node

type Menu []string

type Node struct {
	Name     string
	SubNodes []*Node
	Command  string
	Data     string
}

func makeNodeShort(name string, subNodes []*Node) *Node {
	return &Node{name, subNodes, name, ""}
}

func makeNodeLong(name string, subNodes []*Node, command, data string) *Node {
	return &Node{name, subNodes, name, data}
}

func UberMenu() *Node {
	node := makeNodeLong("Main menu",
		[]*Node{
			appsMenu(),
			historyMenu(),
			//gitMenu(),
			//gitHistoryMenu(),
			//fileManagerMenu(),
			//controlMenu(),
		},
		"", "")
	return node

}

var menuData = `
[
"!arc list",
"!git status",
"git add",
"!!git commit",
"!ls -gGh"
]`

var myMenu Menu

func configFile() *Node {
	return makeNodeShort("Edit Config", []*Node{})
}

/*
func MailSummaries() [][]string {
	lines := getSummaries(50)
	out := [][]string{}
	for _, v := range lines {
		command := ""
		name := v[0]
		data := v[1]
		out = append(out, []string{name, command, data})
	}
	return out
}
*/

/*
func AddAppNodes(n *Node) *Node {

}
*/

func gitHistoryMenu() *Node {
	node := makeNodeShort("previous git commands", []*Node{})
	addTextNodesFromString(node, goof.Grep("git", goof.QC([]string{"fish", "-c", "history"})))
	return node
}

func gitMenu() *Node {
	node := makeNodeShort("git menu", []*Node{})
	addTextNodesFromString(node, git())
	return node
}
func git() string {
	return `\&ls
	\&lslR
git status
git status --porcelain
git push
git pull
git pull --rebase
git commit \&lslR
git commit .
git rebase
git merge
git stash
git stash apply
git diff
git reset
git reset --hard
git branch -a
git show "!git branch -a"
git merge ""
git add \&lslR
git log
git log shortlog
git log -p
git log --oneline
git log --stat
git log --graph
git log --oneline --decorate
git log --oneline --decorate --graph
git log --oneline --decorate --graph --simplify-by-decoration
git diff --summary
git submodule init
git submodule update --init --recursive
git submodule sync
imapcli status
imapcli list
imapcli read 1
imapcli read 2
imapcli read 3
imapcli read 4
imapcli read 5
!set
`
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
var ed *GlobalConfig
var config UserConfig
var confFile string

var lastKey time.Time

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
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
		ed = NewEditor()
		//Create a text formatter
		form = glim.NewFormatter()

		jsonerr := json.Unmarshal([]byte(menuData), &myMenu)
		if jsonerr != nil {
			fmt.Println(jsonerr)
		}
	}()

	updateCurrentNode(makeNodeShort("Loading", []*Node{}))
	go func() {
		//time.Sleep(1 * time.Second)
		updateCurrentNode(UberMenu())
	}()

	//currentNode = addTextNodesFromStrStrStr(currentNode, MailSummaries())

	//currentNode =
	currentThing = []*Node{getCurrentNode()}
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
		}
	}

	//End Nuklear

	if ui {
		for {

			currentNode, currentThing, result = doui(currentNode, currentThing, result)
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
