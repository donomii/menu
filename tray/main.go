package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/browser"

	"github.com/donomii/goof"
	"github.com/donomii/menu"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
	"github.com/mostlygeek/arp"
	"github.com/skratchdot/open-golang/open"
)

type Config struct {
	Name             string
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
}

type InfoStruct struct {
	Name     string
	Services []Service
}

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
	fmt.Printf("Loaded config: %+v", Configuration)
}

func LoadInfo() {
	fmt.Printf("Loading info")
	data, err := ioutil.ReadFile("config/public_info.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Info)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded info: %+v", Info)
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

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, Configuration.Name)
}

func public_info(w http.ResponseWriter, req *http.Request) {
	out, _ := json.Marshal(Info)
	fmt.Fprintf(w, string(out))
}

func webserver() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/", hello)
	http.HandleFunc("/public_info", public_info)
	http.ListenAndServe(":80", nil)
}

func main() {
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

func addTopLevelMenuItems(m *menu.Node) {
	//AddSub(apps, appMen)
	for _, v := range m.SubNodes {
		p := systray.AddMenuItem(fmt.Sprintf("%v", v.Name), v.Command)
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
	b, _ := ioutil.ReadFile("config/usermenu.json")
	//log.Println("Loaded json:", string(b))
	json.Unmarshal(b, &usermenu)
	//log.Println("unmarshal:", err)
	//log.Printf("reconstructed menu: %+v\n", usermenu)
	return &usermenu
}

func makeNetworkPcMenu(hosts []HostService) (*menu.Node, *menu.Node) {
	out := menu.MakeNodeLong("Network", []*menu.Node{}, "", "")
	global := menu.MakeNodeLong("Global Services", []*menu.Node{}, "", "")
	for _, host := range hosts {
		h := menu.MakeNodeLong(host.Ip, []*menu.Node{}, host.Ip, "")
		for _, port := range host.Ports {
			h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", PortMap()[port], port), nil, fmt.Sprintf("%v://%v:%v/", PortMap()[port], host.Ip, port), ""))
		}
		fmt.Printf("Processing services: %+v\n", host.Services)
		for _, s := range host.Services {
			ip := host.Ip
			if s.Ip != "" {
				ip = s.Ip
			}
			if s.Global {
				global.SubNodes = append(global.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v/", s.Protocol, ip, s.Port), ""))
			} else {
				h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", s.Name, s.Port), nil, fmt.Sprintf("%v://%v:%v/", s.Protocol, ip, s.Port), ""))
			}
		}
		out.SubNodes = append(out.SubNodes, h)

	}
	return out, global
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

func ScanConfig() {
	//var hosts = []HostService{}

	networks := Configuration.Networks
	log.Println("Scanning user defined networks:", networks)
	for _, network := range networks {
		log.Println("Scanning user defined network:", network)
		hosts = append(hosts, scanNetwork(network, scanPorts)...)
	}

}

func uniqueifyHosts() {
	temp := map[string]HostService{}
	for _, v := range hosts {
		temp[v.Ip] = v
	}

	out := HostServiceList{}
	for _, v := range temp {
		out = append(out, v)
	}
	sort.Sort(out)
	hosts = out
}
func ScanPublicInfo() {

	for i, v := range hosts {
		url := fmt.Sprintf("http://%v:%v/public_info", v.Ip, 80)
		fmt.Println("Public info url:", url)
		resp, err := http.Get(url)
		if err == nil {
			fmt.Println("Got response")
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				fmt.Println("Got body")
				var s InfoStruct
				err := json.Unmarshal(body, &s)
				if err == nil {
					fmt.Printf("Unmarshalled body %v", s)
					hosts[i].Services = s.Services
				}
			}
		}
	}

}

func onReady() {
	m := UberMenu()
	hosts = []HostService{}
	ArpScan()
	ScanC()
	ScanConfig()

	uniqueifyHosts()
	ScanPublicInfo()
	netmenu, globalmenu := makeNetworkPcMenu(hosts)
	fmt.Printf("%+v, %v\n", m.SubNodes, m)
	systray.AddMenuItem("UMH", "Universal Menu")

	apps := menu.AppsMenu()
	//	js, err := json.MarshalIndent(apps, "", " ")

	//fmt.Println(err)
	//fmt.Println("\n\n\nApps tree as javascript", string(js))
	//fmt.Printf("\n\n\nApps tree %+v\n\n\n", apps)

	//var appMen *systray.MenuItem
	//appMen = systray.AddMenuItem("Applications", "Applications")
	//addMenuTree(appMen, apps, m)
	m.SubNodes = append(m.SubNodes, apps)

	//var netMen *systray.MenuItem
	//netMen = systray.AddMenuItem("Network Menu", "Network menu")

	//addMenuTree(netMen, netmenu, m)

	m.SubNodes = append(m.SubNodes, netmenu)
	m.SubNodes = append(m.SubNodes, globalmenu)

	//	var userMen *systray.MenuItem
	//	userMen = systray.AddMenuItem("User Menu", "User menu")

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
