package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/donomii/menu"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func UberMenu() *menu.Node {
	node := menu.MakeNodeLong("Main menu",
		[]*menu.Node{
			menu.AppsMenu(),
			//menu.HistoryMenu(),
			//menu.GitMenu(),
			//gitHistoryMenu(),
			//fileManagerMenu(),
			menu.ControlMenu(),
		},
		"", "")
	return node
}

func main() {
	onExit := func() {
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	systray.Run(onReady, onExit)
}

func AddSub(m *menu.Node, parent *systray.MenuItem) {
	fmt.Printf("*****%+v, %v\n", m.SubNodes, m)

	for _, v := range m.SubNodes {
		if len(v.SubNodes) > 0 {
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name, v.Command), v.Command)
			AddSub(v, p)
		} else {
			fmt.Println("Adding submenu item ", v.Name)
			parent.AddSubMenuItem(fmt.Sprintf("%v, %v", v.Name, v.Command), v.Command)
		}
	}
}

func onReady() {
	m := UberMenu()
	fmt.Printf("%+v, %v\n", m.SubNodes, m)
	systray.AddMenuItem("UMH", "Universal Menu")
	subMenuTop := systray.AddMenuItem("Test", "SubMenu Test (top)")
	subMenuMiddle := subMenuTop.AddSubMenuItem("SubMenu - Level 2", "SubMenu Test (middle)")
	subMenuMiddle.AddSubMenuItem("SubMenu - Level 3", "SubMenu Test (bottom)")
	subMenuMiddle.AddSubMenuItem("Panic!", "SubMenu Test (bottom)")
	var appMen *systray.MenuItem
	apps := menu.AppsMenu()
	appMen = systray.AddMenuItem("test", "test")
	AddSub(apps, appMen)
	for _, v := range m.SubNodes {

		if len(v.SubNodes) > 0 {
			fmt.Println("Adding submenu to top ", v.Name)
			p := systray.AddMenuItem(fmt.Sprintf("%v, %v", v.Name, v.Command), v.Command)
			AddSub(v, p)
		} else {
			fmt.Println("Adding to top ", v.Name)
			systray.AddMenuItem(fmt.Sprintf("%v, %v", v.Name, v.Command), v.Command)
		}
		//AddSub()
	}
	//fmt.Printf("%+v\n", menu.Apps())
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("UMH")
	systray.SetTooltip("Universaal Menu")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuitOrig.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("Awesome App")
		systray.SetTooltip("Pretty awesome")
		mChange := systray.AddMenuItem("Change Me", "Change Me")
		mChecked := systray.AddMenuItem("Unchecked", "Check Me")
		mEnabled := systray.AddMenuItem("Enabled", "Enabled")
		// Sets the icon of a menu item. Only available on Mac.
		mEnabled.SetTemplateIcon(icon.Data, icon.Data)

		systray.AddMenuItem("Ignored", "Ignored")

		subMenuTop := systray.AddMenuItem("SubMenu", "SubMenu Test (top)")
		subMenuMiddle := subMenuTop.AddSubMenuItem("SubMenu - Level 2", "SubMenu Test (middle)")
		subMenuBottom := subMenuMiddle.AddSubMenuItem("SubMenu - Level 3", "SubMenu Test (bottom)")
		subMenuBottom2 := subMenuMiddle.AddSubMenuItem("Panic!", "SubMenu Test (bottom)")

		mUrl := systray.AddMenuItem("Open UI", "my home")
		mQuit := systray.AddMenuItem("退出", "Quit the whole app")

		// Sets the icon of a menu item. Only available on Mac.
		mQuit.SetIcon(icon.Data)

		systray.AddSeparator()
		mToggle := systray.AddMenuItem("Toggle", "Toggle the Quit button")
		shown := true
		toggle := func() {
			if shown {
				subMenuBottom.Check()
				subMenuBottom2.Hide()
				mQuitOrig.Hide()
				mEnabled.Hide()
				shown = false
			} else {
				subMenuBottom.Uncheck()
				subMenuBottom2.Show()
				mQuitOrig.Show()
				mEnabled.Show()
				shown = true
			}
		}

		for {
			select {
			case <-mChange.ClickedCh:
				mChange.SetTitle("I've Changed")
			case <-mChecked.ClickedCh:
				if mChecked.Checked() {
					mChecked.Uncheck()
					mChecked.SetTitle("Unchecked")
				} else {
					mChecked.Check()
					mChecked.SetTitle("Checked")
				}
			case <-mEnabled.ClickedCh:
				mEnabled.SetTitle("Disabled")
				mEnabled.Disable()
			case <-mUrl.ClickedCh:
				open.Run("https://www.getlantern.org")
			case <-subMenuBottom2.ClickedCh:
				panic("panic button pressed")
			case <-subMenuBottom.ClickedCh:
				toggle()
			case <-mToggle.ClickedCh:
				toggle()
			case <-mQuit.ClickedCh:
				systray.Quit()
				fmt.Println("Quit2 now...")
				return
			}
		}
	}()
}
