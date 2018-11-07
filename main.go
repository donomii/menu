package main

import (
	"image/color"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-shellwords"

	//"text/scanner"

	"flag"
	"fmt"

	"log"
	"os"

	"github.com/donomii/glim"
	"github.com/donomii/goof"
	"github.com/rivo/tview"

	"github.com/donomii/nucular"
	"github.com/donomii/nucular/rect"

	"github.com/donomii/nucular/label"

	nstyle "github.com/donomii/nucular/style"
)

var demoText = "hi"
var result = ""
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

		list.AddItem("Quit", "Press to exit", 'Q', func() {
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
				result = strings.Join(ls("."), "\n")
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

func updatefn(w *nucular.Window) {

	txtSize := 9.6
	if w.Input().Mouse.Buttons[1].Down {
		//col = color.RGBA{255, 0, 0, 0}
		txtSize = 30
	}
	/*
		for _, v := range w.Input().Keyboard.Keys {
			log.Println("%+v", v.)
		}
	*/
	if w.Input().Keyboard.Text != "" {
		//log.Println(w.Input().Keyboard.Text)
		demoText = demoText + w.Input().Keyboard.Text
	}

	col := color.RGBA{255, 255, 255, 255}
	w.Row(30).Dynamic(1)
	header = "\n" + strings.Join(NodesToStringArray(currentThing[1:]), " ")
	w.Label(header, "LC")
	w.Row(30).Dynamic(1)
	w.LabelColored(result, "LC", col)
	/*
		img, _ := glim.DrawStringRGBA(txtSize, col, "Hello again", "f1.ttf")
		newH := img.Bounds().Max.Y
		w.Row(newH).Dynamic(1)
		w.Image(img)
		img2, W, H := glim.GFormatToImage(img, nil, 0, 0)
		img2 = glim.MakeTransparent(img2, color.RGBA{0, 0, 0, 0})
		img3 := glim.Rotate270(W, H, img2)
		img4 := glim.ImageToGFormatRGBA(H, W, img3)
		img5 := img4
		w.Image(img5)
		w.Cmds().DrawImage(rect.Rect{50, 100, 200, 200}, img5)
	*/

	for _, vv := range currentNode.SubNodes {
		//node := vv.SubNodes[i]
		name := vv.Name
		v := vv
		if w.Button(label.T(name), false) {
			if !strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "&") {
				currentThing = append(currentThing, v)
				currentNode = v
			} else {

				log.Println("Running command", name)
				if strings.HasPrefix(name, "!") {

					//It's a shell command

					cmd := name[1:]
					result = goof.Command("/bin/sh", []string{"-c", cmd})
					result = goof.Command("cmd", []string{"/c", cmd})
				}

				if strings.HasPrefix(name, "&") {

					//It's an internal command

					cmd := name[1:]
					if cmd == "lslR" {
						result = strings.Join(lslR("."), "\n")
					}
					if cmd == "ls" {
						result = strings.Join(ls("."), "\n")
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

	if w.Button(label.T("Run your command"), false) {

		result = goof.Command("cmd", []string{"/c", strings.Join(NodesToStringArray(currentThing[1:]), " ")})

		//})
		//textView.SetText(result)
	}

	if w.Button(label.T("Run your interactive command"), false) {

		//result = doQC(NodesToStringArray(currentThing[1:]))
		goof.QCI(NodesToStringArray(currentThing[1:]))

		//textView.SetText(result)
		//app.Run()
	}
	if w.Button(label.T("Change directory"), false) {

		//result = doQC(NodesToStringArray(currentThing[1:]))
		path := strings.Join(NodesToStringArray(currentThing[1:]), "/")
		os.Chdir(path)
		currentNode = makeStartNode()
		currentThing = []*Node{currentNode}

		//textView.SetText(result)
		//app.Run()
	}

	if w.Button(label.T("Go back"), false) {
		//app.Stop()
		if len(currentThing) > 1 {
			currentNode = currentThing[len(currentThing)-2]
			currentThing = currentThing[:len(currentThing)-1]
			//header.SetText(strings.Join(NodesToStringArray(currentThing), " "))
			//list.Clear()
			//populateList(list)
		}
	}
	if w.Button(label.T("Press to exit"), false) {

		fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
		app.Stop()
		os.Exit(0)
	}

	w.Label(result, "LC")

	f := glim.NewFormatter()
	f.Colour = &color.RGBA{255, 255, 255, 255}
	f.FontSize = txtSize
	nw := 1200
	nh := 800
	buff := make([]byte, nw*nh*4)

	//glim.RenderTokenPara(f, 0, 0, 10, 10, nw, nh, nw, nh, 1, 1, buff, tokens, true, true, false)
	//buff2 := glim.Rotate270(nw, nh, buff)
	//nw, nh = nh, nw
	//glim.DumpBuff(buff,uint(nw),uint(nh))
	buff = glim.FlipUp(nw, nh, buff)
	tt := glim.ImageToGFormatRGBA(nw, nh, buff)
	w.Cmds().DrawImage(rect.Rect{0, 0, nw, nh}, tt)
	//log.Printf("%+v", w.Input())

}

func makeStartNode() *Node {
	n := &Node{"Start", []*Node{}}
	addTextNodes(n, git())
	return n
}

func main() {
	runtime.GOMAXPROCS(2)
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()

	currentNode = makeStartNode()

	//    currentNode = addHistoryNodes()

	//currentNode = addTextNodes(currentNode,grep("git", doCommand("fish", []string{"-c", "history"})))
	currentThing = []*Node{currentNode}
	//result := ""
	wnd := nucular.NewMasterWindow(0, "MyWindow", updatefn)
	var theme nstyle.Theme = nstyle.DarkTheme
	scaling := 0.9
	if runtime.GOOS == "darwin" {
		scaling = 1.8
	}
	wnd.SetStyle(nstyle.FromTheme(theme, scaling))
	wnd.Main()
	if ui {
		for {

			currentNode, currentThing, result = doui(currentNode, currentThing, result)
		}
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
	startNode := Node{"Start", []*Node{}}
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
