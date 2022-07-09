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
	"sync"
	"syscall"
	"time"

	//"github.com/donomii/menu"

	menu ".."
	tn "../traynetwork"
	"github.com/donomii/goof"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
)

var noScan bool

var setStatus func(string)

var netmenu2 *systray.MenuItem
var netEntities map[string]tn.HostService

func LoadInfo() {

	fmt.Println("Loading info")
	data, err := ioutil.ReadFile("config/public_info.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &tn.Info)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded info: %+v\n", tn.Info)
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
	tn.LoadConfig()
	for _, host := range tn.Configuration.KnownPeers {
		tn.Hosts = append(tn.Hosts, &tn.HostService{Ip: host, Name: "UncontactablePeer", Ports: []uint{16002}, LastSeen: time.Now()})
		log.Printf("Added known peer %v\n", host)
	}
	LoadInfo()
	ti := time.Second * time.Duration(tn.Configuration.PeerUpdateInterval)
	fmt.Printf("Updating peers every %v seconds\n", ti.Seconds())
	go tn.Webserver(tn.Configuration.HttpPort, tn.Configuration.StartPagePort)
	onExit := func() {
		//now := time.Now()
		//ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
		fmt.Println("done")
	}

	go func() {

		for {
			log.Println("Sending hosts list to peers")

			//Print known hosts
			for _, v := range tn.Hosts {
				fmt.Printf("%+v\n", v.Ip)
			}

			js, _ := json.Marshal(tn.Hosts)
			log.Print(string(js))
			tn.UpdatePeers()
			time.Sleep(ti)

		}
	}()

	for {
		systray.Run(onReady, onExit)
		log.Println("Reloading")
	}
}

var netEntitiesLock sync.Mutex

func AddNetworkNode(entry, service string, port uint) {
	netEntitiesLock.Lock()
	defer netEntitiesLock.Unlock()
	fmt.Println("Adding network node", entry)
	if netEntities == nil {
		netEntities = make(map[string]tn.HostService)
	}
	//var mitem *systray.MenuItem
	if _, ok := netEntities[entry]; !ok {
		ent := tn.HostService{
			Name: entry,
		}

		netEntities[entry] = ent
		//mitem = netmenu2.AddSubMenuItemCheckbox(entry, fmt.Sprintf("network %v", entry), false)

		//ent.MenuItem = mitem
	}
	if ent, ok := netEntities[entry]; ok {
		ent.Services = append(ent.Services, tn.Service{
			Name:        service,
			Port:        int(port),
			Protocol:    "tcp",
			Description: "",
			Global:      false,
		})
		//mitem = ent.MenuItem
		//mitem.AddSubMenuItemCheckbox(service, fmt.Sprintf("port %v", port), false)
	}
}

func AddSub(m *menu.Node, parent *systray.MenuItem) {
	//fmt.Printf("*****%+v, %v\n", m.SubNodes, m)

	for _, v := range m.SubNodes {
		if len(v.SubNodes) > 0 {
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), "")
			AddSub(v, p)
		} else {
			//fmt.Printf("Adding submenu item \"%+v\"\n", v)
			p := parent.AddSubMenuItem(fmt.Sprintf("%v", v.Name), "")
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
		p := systray.AddMenuItem(fmt.Sprintf("%v", v.Name), "")
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
	tn.Hosts = []*tn.HostService{}
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

	usermenu := tn.MakeUserMenu()
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
			ForkExec(exe, os.Args, procAttr)
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
		netmenu2 = systray.AddMenuItem("Network", "Network")
		//Start the network scan after the menu is fully loaded, otherwise we can run out of file
		//handles and not be able to open the config files
		go func() {

			if !noScan {
				tn.PortsToScan = append(tn.PortsToScan, tn.Configuration.HttpPort, tn.Configuration.StartPagePort)
				//tn.ArpScan()
				//tn.ScanC()
				tn.ScanConfig()
				tn.ScanPublicInfo()
				tn.UniqueifyHosts()

			} else {
				fmt.Println("Network scan disabled")
			}

			netmenu, globalmenu := tn.MakeNetworkPcMenu(tn.Hosts)

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
