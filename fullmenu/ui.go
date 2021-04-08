// ui.go
package main

import (
	"os"
	"strings"

	//"text/scanner"

	"fmt"

	"github.com/gdamore/tcell"

	//"github.com/donomii/glim"
	"github.com/donomii/goof"
	"github.com/rivo/tview"
)

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

			cmd := currentNode.Command[1:]
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
			execNode := makeNodeShort("Exec", []*Node{})
			addTextNodesFromString(execNode, result)
			currentNode = execNode
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
