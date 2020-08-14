// main
package main

import (
	"fmt"
	"image/color"

	"github.com/donomii/menu"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func main() {
	fmt.Println("Hello World!")
	doGui()
}

const preferenceCurrentTab = "currentTab"

func doGui() {
	a := app.NewWithID("com.praeceptamachinae.com")
	a.SetIcon(theme.FyneLogo())

	w := a.NewWindow("Menu")
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() { fmt.Println("Menu New") }),
		// a quit item will be appended to our first menu
	)))
	w.SetMaster()

	options := [][]string{[]string{"Hello"}, []string{"Goodbye"}}

	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Repos", theme.ViewFullScreenIcon(), DialogScreen(w, a, options)))
	tabs.SetTabLocation(widget.TabLocationLeading)
	tabs.SelectTabIndex(a.Preferences().Int(preferenceCurrentTab))
	w.SetContent(tabs)
	w.Resize(fyne.NewSize(800, 600))

	w.ShowAndRun()
	a.Preferences().SetInt(preferenceCurrentTab, tabs.CurrentTabIndex())
}

func makeCell() fyne.CanvasObject {
	rect := canvas.NewRectangle(&color.RGBA{128, 128, 128, 255})
	rect.SetMinSize(fyne.NewSize(30, 30))
	return rect
}

var buttons []fyne.CanvasObject

func iota(max int) []int {
	a := make([]int, max)
	for i := range a {
		a[i] = i
	}
	return a
}
func DialogScreen(win fyne.Window, a fyne.App, repos [][]string) fyne.CanvasObject {

	buttons = []fyne.CanvasObject{}

	entry := newEnterEntry()
	entry.SetPlaceHolder("Search")
	entry.Focused()

	//buttons = append(buttons, entry)

	for _, _ = range iota(5) {

		b := widget.NewButton("***", func() {
		})
		buttons = append(buttons, b)
	}

	co := []fyne.CanvasObject{entry}
	co = append(co, buttons...)
	dialogs := widget.NewGroup("Repositories", co...)
	windows := widget.NewVBox(dialogs)
	entry.Focused()
	return fyne.NewContainerWithLayout(layout.NewAdaptiveGridLayout(2), windows)
}

type enterEntry struct {
	widget.Entry
}

func (e *enterEntry) onEnter() {
	fmt.Println(e.Entry.Text)
	//Activate button 1
	e.Entry.SetText("")
}

func newEnterEntry() *enterEntry {
	entry := &enterEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *enterEntry) KeyDown(key *fyne.KeyEvent) {
	fmt.Print(e.Text)
	pred := menu.Predict([]byte(e.Text))
	for i, v := range buttons {
		if i < len(pred) {
			v.(*widget.Button).SetText(pred[i])
		} else {
			v.(*widget.Button).SetText("***")
		}
	}
	fmt.Printf("%+v", pred)
	switch key.Name {
	case fyne.KeyReturn:
		e.onEnter()
	default:
		e.Entry.KeyDown(key)
	}

}
