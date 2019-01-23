package main

import (
	"encoding/json"
	"runtime"
	"strings"

	"golang.org/x/image/font/gofont/goregular"

	//"unsafe"

	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"

	"github.com/mattn/go-shellwords"

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

var currentNode *Node
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

var menuData = `
[
"!arc list",
"!git status",
"git add",
"!!git commit",
"!ls -gGh"
]`

var myMenu Menu

func NodesToStringArray(ns []*Node) []string {
	var out []string
	for _, v := range ns {
		out = append(out, v.Name)

	}
	return out

}

func Apps() [][]string {
	lines := strings.Split(goof.QC([]string{"ls", "/Applications"}), "\n")
	out := [][]string{}
	for _, v := range lines {
		name := strings.TrimSuffix(v, ".app")
		command := fmt.Sprintf("!open \"/Applications/%v\"", v)
		out = append(out, []string{name, command})
	}
	return out
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

func makeStartNode() *Node {
	n := makeNodeShort("Command:", []*Node{})

	return n
}

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

var winWidth = 900
var winHeight = 900

func main() {

	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()

	//Create a text formatter
	form = glim.NewFormatter()

	jsonerr := json.Unmarshal([]byte(menuData), &myMenu)
	if jsonerr != nil {
		fmt.Println(jsonerr)
	}

	//myMenu = Apps()
	//fmt.Println("Menu: ", myMenu)

	currentNode = makeStartNode()
	//currentNode = addTextNodesFromCommands(currentNode, myMenu)

	//currentNode = addTextNodesFromStrStrStr(currentNode, MailSummaries())

	addTextNodesFromString(currentNode, git())
	//    currentNode = addHistoryNodes()

	//currentNode = addTextNodes(currentNode,grep("git", doCommand("fish", []string{"-c", "history"})))
	currentThing = []*Node{currentNode}
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

func (n *Node) String() string {
	return n.Name
}

func (n *Node) ToString() string {
	return n.Name
}

func findNode(n *Node, name string) *Node {
	if n == nil {
		return n
	}
	for _, v := range n.SubNodes {
		if v.Name == name {
			return v
		}
	}
	return nil

}

func addFileNodes() {

}

func addHistoryNodes() *Node {
	src := goof.Command("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := makeStartNode()
	for _, l := range lines {
		currentNode := startNode
		/*
				var s scanner.Scanner
				s.Init(strings.NewReader(l))
				s.Filename = "example"
				for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			        text := s.TokenText()
					fmt.Printf("%s: %s\n", s.Position, text)
			        if findNode(currentNode, text) == nil {
			            newNode := Node{text, []*Node{}}
			            currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
			            currentNode = &newNode
			        } else {
			            currentNode = findNode(currentNode, text)
			        }
		*/
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}

		}
	}
	return startNode
}

func addTextNodesFromString(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	return addTextNodesFromStringList(startNode, lines)
}

func appendNewNodeShort(text string, aNode *Node) *Node {
	newNode := makeNodeShort(text, []*Node{})
	aNode.SubNodes = append(aNode.SubNodes, newNode)
	return aNode
}

func addTextNodesFromStringList(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}
		}
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromCommands(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		appendNewNodeShort(l, startNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], ""}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromStrStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], l[2]}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func dumpTree(n *Node, indent int) {
	fmt.Printf("%*s%s\n", indent, "", n.Name)
	for _, v := range n.SubNodes {
		dumpTree(v, indent+1)
	}

}
