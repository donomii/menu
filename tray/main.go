package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/donomii/goof"
	"github.com/donomii/menu"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
	"github.com/mostlygeek/arp"
	"github.com/skratchdot/open-golang/open"
)

func UberMenu() *menu.Node {
	node := menu.MakeNodeLong("Main menu",
		[]*menu.Node{
		//menu.AppsMenu(),
		//menu.HistoryMenu(),
		//menu.GitMenu(),
		//gitHistoryMenu(),
		//fileManagerMenu(),
		//menu.ControlMenu(),
		},
		"", "")
	return node
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}
func webserver() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/", hello)
	http.ListenAndServe(":80", nil)
}

func main() {
	//go ScanAll()
	arp.AutoRefresh(1 * time.Second)
	go webserver()
	onExit := func() {
		//now := time.Now()
		//ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
		fmt.Println("done")
	}

	for {
		systray.Run(onReady, onExit)
	}
}

func AddSub(m *menu.Node, parent *systray.MenuItem) {
	//fmt.Printf("*****%+v, %v\n", m.SubNodes, m)

	for _, v := range m.SubNodes {
		if len(v.SubNodes) > 0 {
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
			AddSub(v, p)
		} else {
			fmt.Printf("Adding submenu item \"%+v\"\n", v)
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
			go func(v *menu.Node, p *systray.MenuItem) {
				for {
					<-p.ClickedCh
					fmt.Println("Clicked2", v.Name)
					fmt.Println("Clicked2", v.Command)
					if strings.HasPrefix(v.Command, "exec://") {
						cmd := strings.TrimPrefix(v.Command, "exec://")
						log.Println("Executing", cmd)
						go goof.QC([]string{cmd})
					} else if strings.HasPrefix(v.Command, "http") {
						log.Println("Opening", v.Command, "in browser")
						browser.OpenURL(v.Command)
					} else if goof.Exists(v.Command) {
						log.Println("Opening", v.Command, "as document")
						goof.QC([]string{"rundll32.exe", "url.dll,FileProtocolHandler", v.Command})
					} else if strings.HasPrefix(v.Command, "shell://") {
						cmd := strings.TrimPrefix(v.Command, "shell://")
						log.Println("Opening", cmd, "as shell command")
						go goof.QC([]string{"cmd", "/K", cmd})
					}
				}
			}(v, p)
		}
	}
}

func addMenuTree(appMen *systray.MenuItem, apps, m *menu.Node) {
	AddSub(apps, appMen)
	for _, v := range m.SubNodes {
		p := systray.AddMenuItem(fmt.Sprintf("%v, %v", v.Name, v.Command), v.Command)
		go func(v *menu.Node) {
			for {
				<-p.ClickedCh
				fmt.Println("Clicked2", v.Name)
			}
		}(v)
		if len(v.SubNodes) > 0 {
			fmt.Println("Adding submenu ", v.Name)
			AddSub(v, p)
		} else {
			fmt.Println("Adding menu item", v.Name)
		}

	}
}

func makeUserMenu() *menu.Node {
	var usermenu menu.Node
	b, _ := ioutil.ReadFile("usermenu.json")
	log.Println("Loaded json:", string(b))
	err := json.Unmarshal(b, &usermenu)
	log.Println("unmarshal:", err)
	log.Printf("reconstructed menu: %+v\n", usermenu)
	return &usermenu
}

func makeNetworkPcMenu(hosts []HostService) *menu.Node {
	out := menu.MakeNodeLong("Network", []*menu.Node{}, "", "")
	for _, host := range hosts {
		h := menu.MakeNodeLong(host.Ip, []*menu.Node{}, host.Ip, "")
		for _, port := range host.Ports {
			h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", PortMap()[port], port), nil, fmt.Sprintf("%v://%v:%v/", PortMap()[port], host.Ip, port), ""))
		}
		out.SubNodes = append(out.SubNodes, h)

	}
	return out
}

var hosts = []HostService{}
var scanPorts = []int{137, 138, 139, 445, 80, 443, 20, 21, 22, 23, 25, 53, 3000, 8000, 8001, 8080, 8081, 8008}

func ArpScan() {
	//var hosts = []HostService{}

	keys := []string{}
	for k, _ := range arp.Table() {
		keys = append(keys, k)
	}
	log.Printf("Scanning %v\n", keys)
	hosts = append(hosts, scanIps(keys, scanPorts)...)

}
func ScanC() {
	//var hosts = []HostService{}

	ips := goof.AllIps()

	classB := map[string]bool{}
	for _, ip := range ips {
		hosts = append(hosts, scanNetwork(ip+"/24", scanPorts)...)
		bits := strings.Split(ip, ".")
		b := bits[0] + "." + bits[1] + ".0.0"
		classB[b] = true
	}
	/*
		for ip, _ := range classB {
			hosts = append(hosts, scanNetwork(ip+"/16")...)
		}
	*/
}

func onReady() {
	m := UberMenu()
	hosts = []HostService{}
	ArpScan()
	ScanC()
	netmenu := makeNetworkPcMenu(hosts)
	fmt.Printf("%+v, %v\n", m.SubNodes, m)
	systray.AddMenuItem("UMH", "Universal Menu")
	var appMen *systray.MenuItem
	apps := menu.AppsMenu()
	//	js, err := json.MarshalIndent(apps, "", " ")

	//fmt.Println(err)
	//fmt.Println("\n\n\nApps tree as javascript", string(js))
	//fmt.Printf("\n\n\nApps tree %+v\n\n\n", apps)
	appMen = systray.AddMenuItem("Applications", "Applications")
	addMenuTree(appMen, apps, m)

	var userMen *systray.MenuItem
	userMen = systray.AddMenuItem("User Menu", "User menu")

	usermenu := makeUserMenu()
	addMenuTree(userMen, usermenu, m)

	var netMen *systray.MenuItem
	netMen = systray.AddMenuItem("Network Menu", "Network menu")

	addMenuTree(netMen, netmenu, m)

	mQuitOrig := systray.AddMenuItem("Reload", "Reload menu")
	go func() {
		<-mQuitOrig.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("UMH")
		systray.SetTooltip("Universal Menu")
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
