package main

import (
	"encoding/json"
	"io/ioutil"
	"runtime"

	"github.com/donomii/menu"

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

var form *glim.FormatParams
var demoText = "hi"
var displaySplit string = "None"
var result = ""
var EditStr = `lalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalala`
var EditBytes []byte
var tokens [][]string

var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var app *tview.Application
var workerChan chan string
var currentNodeLock sync.Mutex

var currentNode *menu.Node

func updateCurrentNode(n *menu.Node) {
	currentNodeLock.Lock()
	currentNode = n
	currentNodeLock.Unlock()
}

func getCurrentNode() *menu.Node {
	currentNodeLock.Lock()
	n := currentNode
	currentNodeLock.Unlock()
	return n
}

var currentThing []*menu.Node

type Menu []string

type Node struct {
	Name     string
	SubNodes []*Node
	Command  string
	Data     string
}

func UberMenu() *menu.Node {
	node := menu.MakeNodeLong("Main menu",
		[]*menu.Node{
			menu.AppsMenu(),
			menu.HistoryMenu(),
			//menu.GitMenu(),
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

func configFile() *menu.Node {
	return menu.MakeNodeShort("Edit Config", []*menu.Node{})
}

func MailSummaries() [][]string {
	username := goof.CatFile("username")
	password := goof.CatFile("password")
	lines := menu.GetSummaries(4, username, password)
	out := [][]string{}
	for _, v := range lines {
		command := ""
		name := v[0]
		data := v[1]
		out = append(out, []string{name, command, data})
	}
	return out
}

/*
func AddAppNodes(n *Node) *Node {

}
*/

func gitHistoryMenu() *menu.Node {
	node := menu.MakeNodeShort("previous git commands", []*menu.Node{})
	str, _ := goof.QC([]string{"fish", "-c", "history"})
	menu.AddTextNodesFromString(node, goof.Grep("git", str))
	return node
}

func gitMenu() *menu.Node {
	node := menu.MakeNodeShort("git menu", []*menu.Node{})
	menu.AddTextNodesFromString(node, git())
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

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
	log.Println("Locked to main thread")
}

func main() {
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

	toml.Decode(string(configBytes), &config)
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()

	go func() {
		ed = NewEditor()
		//Create a text formatter
		form = glim.NewFormatter()

		jsonerr := json.Unmarshal([]byte(menuData), &myMenu)
		if jsonerr != nil {
			fmt.Println(jsonerr)
		}
	}()

	updateCurrentNode(menu.MakeNodeShort("Loading", []*menu.Node{}))
	go func() {
		//time.Sleep(1 * time.Second)
		updateCurrentNode(UberMenu())
	}()

	currentNode = menu.AddTextNodesFromStrStrStr(currentNode, MailSummaries())

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
	log.Printf("glfw: created window %dx%d", width, height)

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	gl.Viewport(0, 0, int32(width-1), int32(height-1))

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	/*data, err := ioutil.ReadFile("FreeSans.ttf")
	if err != nil {
		panic("Could not find file")
	}*/

	sansFont := nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 16, nil)
	// sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	}

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
