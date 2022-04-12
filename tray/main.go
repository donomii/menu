package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"

	//"github.com/donomii/menu"

	"github.com/donomii/goof"
	"github.com/donomii/menu"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
)

var noScan bool

type Config struct {
	HttpPort         uint
	StartPagePort    uint
	Name             string
	MaxUploadSize    uint
	Networks         []string
	ArpCheckInterval int
}

var Configuration Config

type Service struct {
	Name        string
	Ip          string
	Port        int
	Protocol    string
	Description string
	Global      bool
	Path        string
}

type InfoStruct struct {
	Name     string
	Services []Service
}

var setStatus func(string)

var Info InfoStruct

func LoadConfig() {
	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Configuration)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded config: %+v\n", Configuration)
}

func LoadInfo() {
	fmt.Println("Loading info")
	data, err := ioutil.ReadFile("config/public_info.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Info)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded info: %+v\n", Info)
}

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

func main() {
	flag.BoolVar(&noScan, "no-scan", false, "Don't do network scan at start")
	flag.Parse()
	baseDir := goof.ExecutablePath()
	os.Chdir(baseDir)
	//go ScanAll()
	LoadConfig()
	LoadInfo()

	go webserver()
	onExit := func() {
		//now := time.Now()
		//ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
		fmt.Println("done")
	}

	for {
		systray.Run(onReady, onExit)
		log.Println("Reloading")
	}
}

func AddSub(m *menu.Node, parent *systray.MenuItem) {
	//fmt.Printf("*****%+v, %v\n", m.SubNodes, m)

	for _, v := range m.SubNodes {
		if len(v.SubNodes) > 0 {
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
			AddSub(v, p)
		} else {
			//fmt.Printf("Adding submenu item \"%+v\"\n", v)
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
			go func(v *menu.Node, p *systray.MenuItem) {
				for {
					<-p.ClickedCh
					fmt.Println("Clicked2", v.Name)
					fmt.Println("Clicked2", v.Command)
					menu.Activate(v.Command)
				}
			}(v, p)
		}
	}
}

func addTopLevelMenuItems(m *menu.Node) {
	//AddSub(apps, appMen)
	for _, v := range m.SubNodes {
		p := systray.AddMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
		go func(v *menu.Node) {
			for {
				<-p.ClickedCh
				fmt.Println("Clicked top level", v.Name)
				menu.Activate(v.Command)
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
	b, _ := ioutil.ReadFile("config/usermenu.json")
	json.Unmarshal(b, &usermenu)
	return &usermenu
}

func makeNetworkPcMenu(hosts []HostService) (*menu.Node, *menu.Node) {
	out := menu.MakeNodeLong("Network", []*menu.Node{}, "", "")
	global := menu.MakeNodeLong("Global Services", []*menu.Node{}, "", "")
	for _, host := range hosts {
		h := menu.MakeNodeLong(host.Ip+"/"+host.Name, []*menu.Node{}, host.Ip, "")
		for _, port := range host.Ports {
			protocol := "http"
			if port == 443 {
				protocol = "https"
			}
			h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", PortMap()[int(port)], port), nil, fmt.Sprintf("%v://%v:%v/", protocol, host.Ip, port), ""))
		}
		fmt.Printf("Processing services: %+v\n", host.Services)
		for _, s := range host.Services {
			ip := host.Ip
			if s.Ip != "" {
				ip = s.Ip
			}
			protocol := "http"
			if s.Port == 443 {
				protocol = "https"
			}
			if !strings.HasPrefix(s.Path, "/") {
				s.Path = "/" + s.Path
			}
			if s.Global {
				global.SubNodes = append(global.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v%v", s.Protocol, ip, s.Port, s.Path), ""))
			} else {
				h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v%v", protocol, ip, s.Port, s.Path), ""))
			}
		}
		out.SubNodes = append(out.SubNodes, h)

	}
	return out, global
}

func RecallMenu() *menu.Node {
	m := menu.Recall()
	out := menu.MakeNodeLong("Recall", []*menu.Node{}, "", "")
	for _, entry := range m {
		h := menu.MakeNodeLong(entry[0], []*menu.Node{}, entry[1], "")
		out.SubNodes = append(out.SubNodes, h)
	}
	return out
}

func trim(s string) string {
	out := strings.TrimSpace(s)
	return out
}

func listWifi() []string {
	//Macosx
	str, _ := goof.QC([]string{"/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport", "scan"})
	wifi_str := trim(str)
	lines := strings.Split(wifi_str, "\n")
	out := []string{}
	for _, l := range lines {
		ll := trim(l)
		bits := strings.Split(ll, " ")
		out = append(out, bits[0])
	}
	return out
}

func makeWifiMenu(ssids []string) *menu.Node {
	out := menu.MakeNodeLong("Wifi", []*menu.Node{}, "", "")
	for _, network := range ssids {
		h := menu.MakeNodeLong(network, []*menu.Node{}, "shell://networksetup -setairportnetwork en0 \""+network+"\" password_goes_here", "")

		out.SubNodes = append(out.SubNodes, h)
	}
	return out
}

func onReady() {
	m := UberMenu()
	hosts = []HostService{}
	netmenus := menu.Node{Name: "Network", SubNodes: []*menu.Node{}}

	fmt.Printf("%+v, %v\n", m.SubNodes, m)
	systray.AddMenuItem("UMH", "Universal Menu")

	var apps *menu.Node
	if runtime.GOOS == "darwin" {
		apps = menu.AppsMenu()
	} else {
		apps = menu.TieredAppsMenu()
	}
	m.SubNodes = append(m.SubNodes, apps)

	m.SubNodes = append(m.SubNodes, makeWifiMenu(listWifi()))

	usermenu := makeUserMenu()
	m.SubNodes = append(m.SubNodes, usermenu)
	m.SubNodes = append(m.SubNodes, usermenu.SubNodes...)
	m.SubNodes = append(m.SubNodes, RecallMenu())
	addTopLevelMenuItems(m)

	mQuitOrig := systray.AddMenuItem("Reload", "Reload menu")
	go func() {
		<-mQuitOrig.ClickedCh

		if runtime.GOOS == "darwin" {
			procAttr := new(syscall.ProcAttr)
			procAttr.Files = []uintptr{0, 1, 2}
			procAttr.Dir = os.Getenv("PWD")
			procAttr.Env = os.Environ()
			exe, _ := os.Executable()
			syscall.ForkExec(exe, os.Args, procAttr)
		}
		systray.Quit()
		fmt.Println("Systray stopped")
	}()

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("UMH")
		systray.SetTooltip("Universal Menu")
		statusMenuItem := systray.AddMenuItem("-------------", " --------------")
		setStatus = func(status string) {
			statusMenuItem.SetTitle(status)
		}

		//mChecked := systray.AddMenuItem("Unchecked", "Check Me")

		mQuit := systray.AddMenuItem("退出", "Quit the whole app")

		// Sets the icon of a menu item. Only available on Mac.
		mQuit.SetIcon(icon.Data)

		//Start the network scan after the menu is fully loaded, otherwise we can run out of file
		//handles and not be able to open the config files
		go func() {

			if !noScan {

				ArpScan()
				ScanC()
				ScanConfig()

				uniqueifyHosts()
				ScanPublicInfo()
			} else {
				fmt.Println("Network scan disabled")
			}

			netmenu, globalmenu := makeNetworkPcMenu(hosts)

			netmenus.SubNodes = append(netmenus.SubNodes, netmenu)
			netmenus.SubNodes = append(netmenus.SubNodes, globalmenu)
			addTopLevelMenuItems(&netmenus)
		}()

		for {
			select {
			case <-statusMenuItem.ClickedCh:
				statusMenuItem.SetTitle("----------------")
				statusMenuItem.SetIcon(icon.Data)
				/*
					case <-mChecked.ClickedCh:
						if mChecked.Checked() {
							mChecked.Uncheck()
							mChecked.SetTitle("Unchecked")
						} else {
							mChecked.Check()
							mChecked.SetTitle("Checked")
						}
				*/
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}
