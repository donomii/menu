package main

import (
	"io/ioutil"
	"runtime"
	"strings"

	"golang.org/x/image/font/gofont/goregular"

	//"unsafe"

	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-shellwords"

	//"text/scanner"

	"flag"
	"fmt"

	"log"
	"os"

	//"github.com/donomii/glim"
	"github.com/donomii/goof"
	"github.com/rivo/tview"
)

var demoText = "hi"
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

func NodesToStringArray(ns []*Node) []string {
	var out []string
	for _, v := range ns {
		out = append(out, v.Name)

	}
	return out

}
func doui(cN *Node, cT []*Node, extraText string) (currentNode *Node, currentThing []*Node, result string) {
	currentNode = cN
	currentThing = cT

	//box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	app = tview.NewApplication()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textView.SetText(extraText)

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	footer.SetText("lalalala")

	newPrimitive := func(text string) *tview.TextView {
		p := tview.NewTextView()
		p.SetTextAlign(tview.AlignCenter).
			SetText(text)
		p.SetChangedFunc(func() {
			app.Draw()
		})
		return p
	}
	header := newPrimitive("")
	header.SetText(strings.Join(NodesToStringArray(currentThing), " "))
	header.SetTextColor(tcell.ColorRed)

	list := tview.NewList()
	populateList := func(list *tview.List) { os.Exit(0) }
	extendList := func(list *tview.List) {
		list.AddItem("Run", "Run your text", 'R', func() {
			//app.Stop()
			//app.Suspend(func() {
			result = goof.Command("/bin/sh", []string{"-c", strings.Join(NodesToStringArray(currentThing[1:]), " ")})
			//})
			textView.SetText(result)
			//app.Run()
			//app.Draw()
		})
		list.AddItem("Run Interactive", "Run your text", 'R', func() {
			//app.Stop()
			app.Suspend(func() {
				//result = doQC(NodesToStringArray(currentThing[1:]))
				goof.QCI(NodesToStringArray(currentThing[1:]))
			})
			textView.SetText(result)
			//app.Run()
		})
		list.AddItem("Back", "Go back", 'B', func() {
			//app.Stop()
			if len(currentThing) > 1 {
				currentNode = currentThing[len(currentThing)-2]
				currentThing = currentThing[:len(currentThing)-1]
				header.SetText(strings.Join(NodesToStringArray(currentThing), " "))
				list.Clear()
				populateList(list)
			}
		})

		list.AddItem("Quit", "exit", 'Q', func() {
			fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
			app.Stop()
			os.Exit(0)
		})
		app.Draw()
	}

	populateList = func(list *tview.List) {
		list.Clear()
		result = ""
		if strings.HasPrefix(currentNode.Name, "!") {

			//It's a shell command

			cmd := currentNode.Name[1:]
			result = goof.Command("/bin/sh", []string{"-c", cmd})
		}

		if strings.HasPrefix(currentNode.Name, "&") {

			//It's an internal command

			cmd := currentNode.Name[1:]
			if cmd == "lslR" {
				result = strings.Join(goof.LslR("."), "\n")
			}
			if cmd == "ls" {
				result = strings.Join(goof.Ls("."), "\n")
			}
		}

		if result != "" {
			execNode := Node{"Exec", []*Node{}}
			addTextNodes(&execNode, result)
			currentNode = &execNode
		}
		for i, vv := range currentNode.SubNodes {
			//node := vv.SubNodes[i]
			name := vv.Name
			v := vv
			list.AddItem(name, name, goof.ToChar(i), func() {
				if !strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "&") {
					currentThing = append(currentThing, v)
				}
				currentNode = v

				header.SetText("\n" + strings.Join(NodesToStringArray(currentThing[1:]), " "))
				list.Clear()
				populateList(list)
				//app.Stop()
			})
		}
		extendList(list)
	}

	populateList(list)

	//menu := newPrimitive("Menu")
	//sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(3, 0, 2).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(footer, 2, 0, 1, 2, 0, 0, false)

	/*
		        grid.AddItem(menu, 0, 0, 1, 3, 0, 0, false).
		        AddItem(list, 1, 0, 1, 3, 0, 0, true).
				AddItem(sideBar, 0, 0, 1, 3, 0, 0, false)
	*/

	grid.AddItem(list, 1, 0, 1, 1, 0, 40, true).
		AddItem(textView, 1, 1, 1, 1, 0, 40, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
	return currentNode, currentThing, result
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

func makeStartNode() *Node {
	n := &Node{"Command:", []*Node{}}
	addTextNodes(n, git())
	return n
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

	currentNode = makeStartNode()

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

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, "Menu", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	nk.NkWindowSetPosition(ctx, "Menu", nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, "Menu", nk.NkVec2(float32(winWidth), float32(winHeight)))

	if update > 0 {
		nk.NkLayoutRowDynamic(ctx, 20, 3)
		{
			nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " "), nk.TextLeft)
			if 0 < nk.NkButtonLabel(ctx, "Undo") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
					currentThing = currentThing[:len(currentThing)-1]
				}
			}
			if 0 < nk.NkButtonLabel(ctx, "Go Back") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
				}
			}
		}
		QuickFileEditor(ctx)

		nk.NkLayoutRowDynamic(ctx, 20, 3)
		{
			nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " "), nk.TextLeft)
			if 0 < nk.NkButtonLabel(ctx, "Run") {
				cmd := strings.Join(NodesToStringArray(currentThing[1:]), " ")
				result = goof.Command("cmd", []string{"/c", cmd})
				result = result + goof.Command("/bin/sh", []string{"-c", cmd})
			}

			if 0 < nk.NkButtonLabel(ctx, "Run interactive") {
				goof.QCI(NodesToStringArray(currentThing[1:]))

			}
		}
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		{

			//pf := nk.NewPluginFilterRef(unsafe.Pointer(&nk.NkFilterDefault))

			size := nk.NkVec2(nk.NkWidgetWidth(ctx), 400)
			if nk.NkComboBeginColor(ctx, state.bgColor, size) > 0 {
				nk.NkLayoutRowDynamic(ctx, 120, 1)
				state.bgColor = nk.NkColorPicker(ctx, state.bgColor, nk.ColorFormatRGBA)
				nk.NkLayoutRowDynamic(ctx, 25, 1)
				r, g, b, a := state.bgColor.RGBAi()
				r = nk.NkPropertyi(ctx, "#R:", 0, r, 255, 1, 1)
				g = nk.NkPropertyi(ctx, "#G:", 0, g, 255, 1, 1)
				b = nk.NkPropertyi(ctx, "#B:", 0, b, 255, 1, 1)
				a = nk.NkPropertyi(ctx, "#A:", 0, a, 255, 1, 1)
				state.bgColor.SetRGBAi(r, g, b, a)
				nk.NkComboEnd(ctx)
			}
		}

	}
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func ButtonBox(ctx *nk.Context) {
	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			for _, vv := range currentNode.SubNodes {
				//node := vv.SubNodes[i]
				name := vv.Name
				v := vv
				if nk.NkButtonLabel(ctx, name) > 0 {
					if !strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "&") {
						currentThing = append(currentThing, v)
						currentNode = v
					} else {

						log.Println("Running command", name)
						if strings.HasPrefix(name, "!") {

							//It's a shell command

							cmd := name[1:]
							result = goof.Command("/bin/sh", []string{"-c", cmd})
							result = result + goof.Command("cmd", []string{"/c", cmd})
						}

						if strings.HasPrefix(name, "&") {

							//It's an internal command

							cmd := name[1:]
							if cmd == "lslR" {
								result = strings.Join(goof.LslR("."), "\n")
							}
							if cmd == "ls" {
								result = strings.Join(goof.Ls("."), "\n")
							}
						}

						if result != "" {
							log.Println("Ran command, got result", result)
							execNode := Node{"Exec", []*Node{}}
							addTextNodes(&execNode, result)
							currentNode = &execNode
						}

					}

					//list.Clear()
					//populateList(list)
					//app.Stop()
				}
			}

			if 0 < nk.NkButtonLabel(ctx, "Run command") {
				cmd := strings.Join(NodesToStringArray(currentThing[1:]), " ")
				result = goof.Command("cmd", []string{"/c", cmd})
				result = result + goof.Command("/bin/sh", []string{"-c", cmd})
			}

			if 0 < nk.NkButtonLabel(ctx, "Run command interactively") {
				goof.QCI(NodesToStringArray(currentThing[1:]))

			}
			if 0 < nk.NkButtonLabel(ctx, "Change directory") {
				path := strings.Join(NodesToStringArray(currentThing[1:]), "/")
				os.Chdir(path)
				currentNode = makeStartNode()
				currentThing = []*Node{currentNode}
			}

			if 0 < nk.NkButtonLabel(ctx, "Go back") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
					currentThing = currentThing[:len(currentThing)-1]
				}
			}
			if 0 < nk.NkButtonLabel(ctx, "Exit") {

				fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
				app.Stop()
				os.Exit(0)
			}
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 10, 1)
		{
			results := strings.Split(result, "\n")
			for _, v := range results {
				//nk.NkLabel(ctx, v, nk.WindowBorder)
				if nk.NkButtonLabel(ctx, v) > 0 {
					n := &Node{v, []*Node{}}
					currentThing = append(currentThing, n)

				}
			}

		}
		nk.NkGroupEnd(ctx)
	}

}

func QuickFileEditor(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(winHeight), 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{

			files := goof.Ls(".")
			for _, vv := range files {
				if nk.NkButtonLabel(ctx, vv) > 0 {
					var err error
					EditBytes, err = ioutil.ReadFile(vv)
					log.Println(err)
				}

			}
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)

		//nk.NkLayoutRowStatic(ctx, 100, 100, 3)
		nk.NkLayoutRowDynamic(ctx, float32(winHeight), 1)
		{
			if EditBytes != nil {
				//var lenStr = int32(len(EditBytes))
				//nk.NkEditString(ctx, nk.EditMultiline|nk.EditAlwaysInsertMode, EditBytes, &lenStr, 512, nk.NkFilterAscii) FIXME
				nk.NkLabelWrap(ctx, string(EditBytes))
			}
		}
		nk.NkGroupEnd(ctx)
	}

}

type Node struct {
	Name     string
	SubNodes []*Node
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

func addHistoryNodes() *Node {
	src := goof.Command("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := Node{"Command:", []*Node{}}
	for _, l := range lines {
		currentNode := &startNode
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
				newNode := Node{text, []*Node{}}
				currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
				currentNode = &newNode
			} else {
				currentNode = findNode(currentNode, text)
			}

		}
	}
	return &startNode
}

func addTextNodes(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := Node{text, []*Node{}}
				currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
				currentNode = &newNode
			} else {
				currentNode = findNode(currentNode, text)
			}
		}
	}

	//fmt.Println()
	//fmt.Printf("%+v\n", startNode)
	//dumpTree(startNode, 0)
	return startNode

}

func dumpTree(n *Node, indent int) {
	fmt.Printf("%*s%s\n", indent, "", n.Name)
	for _, v := range n.SubNodes {
		dumpTree(v, indent+1)
	}

}
