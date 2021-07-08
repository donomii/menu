package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	//"github.com/donomii/menu"

	"github.com/donomii/goof"
	"github.com/donomii/menu"

	//".."
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
	scanPorts = append(scanPorts, Configuration.HttpPort)
	scanPorts = append(scanPorts, Configuration.StartPagePort)
	LoadInfo()

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
			h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", PortMap()[int(port)], port), nil, fmt.Sprintf("%v://%v:%v/", PortMap()[int(port)], host.Ip, port), ""))
		}
		fmt.Printf("Processing services: %+v\n", host.Services)
		for _, s := range host.Services {
			ip := host.Ip
			if s.Ip != "" {
				ip = s.Ip
			}
			if s.Global {
				global.SubNodes = append(global.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v/%v", s.Protocol, ip, s.Port, s.Path), ""))
			} else {
				h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v/%v", s.Protocol, ip, s.Port, s.Path), ""))
			}
		}
		out.SubNodes = append(out.SubNodes, h)

	}
	return out, global
}

func trim(s string) string {
	out := strings.TrimSpace(s)
	return out
}

func listWifi() []string {
	//Macosx
	wifi_str := trim(goof.QC([]string{"/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport", "scan"}))
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
	if !noScan {

		ArpScan()
		ScanC()
		ScanConfig()

		uniqueifyHosts()
		ScanPublicInfo()
	}
	fmt.Println("NoScan:", noScan)

	netmenu, globalmenu := makeNetworkPcMenu(hosts)

	fmt.Printf("%+v, %v\n", m.SubNodes, m)
	systray.AddMenuItem("UMH", "Universal Menu")

	var apps *menu.Node
	if runtime.GOOS == "darwin" {
		apps = menu.AppsMenu()
	} else {
		apps = menu.TieredAppsMenu()
	}
	m.SubNodes = append(m.SubNodes, apps)

	m.SubNodes = append(m.SubNodes, netmenu)
	m.SubNodes = append(m.SubNodes, globalmenu)
	m.SubNodes = append(m.SubNodes, makeWifiMenu(listWifi()))

	usermenu := makeUserMenu()
	m.SubNodes = append(m.SubNodes, usermenu)
	m.SubNodes = append(m.SubNodes, usermenu.SubNodes...)
	addTopLevelMenuItems(m)

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
		statusMenuItem := systray.AddMenuItem("-------------", " --------------")
		setStatus = func(status string) {
			statusMenuItem.SetTitle(status)
		}

		//mChecked := systray.AddMenuItem("Unchecked", "Check Me")

		mQuit := systray.AddMenuItem("退出", "Quit the whole app")

		// Sets the icon of a menu item. Only available on Mac.
		mQuit.SetIcon(icon.Data)

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
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}
