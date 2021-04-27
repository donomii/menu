package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/donomii/goof"
	"github.com/donomii/menu"
	"github.com/donomii/menu/tray/icon"
	"github.com/getlantern/systray"
	"github.com/mostlygeek/arp"
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

	keys := []string{}
	for k, _ := range arp.Table() {
		keys = append(keys, k)
	}
	log.Printf("Scanning %v\n", keys)
	hosts = append(hosts, scanIps(keys, scanPorts)...)

}
func ScanC() {

	ips := goof.AllIps()

	classB := map[string]bool{}
	for _, ip := range ips {
		hosts = append(hosts, scanNetwork(ip+"/24", scanPorts)...)
		bits := strings.Split(ip, ".")
		b := bits[0] + "." + bits[1] + ".0.0"
		classB[b] = true
	}
}

func ScanConfig() {
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

	m.SubNodes = append(m.SubNodes, apps)

	m.SubNodes = append(m.SubNodes, netmenu)
	m.SubNodes = append(m.SubNodes, globalmenu)

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
