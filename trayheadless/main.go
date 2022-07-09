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
	"time"

	//"github.com/donomii/menu"

	menu ".."
	tn "../traynetwork"
	"github.com/donomii/goof"
)

var noScan bool

var setStatus func(string)

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
	LoadInfo()
	go func() {

		if !noScan {
			tn.PortsToScan = append(tn.PortsToScan, tn.Configuration.HttpPort, tn.Configuration.StartPagePort)
			//tn.ArpScan()
			//tn.ScanC()
			for _, host := range tn.Configuration.KnownPeers {
				tn.Hosts = append(tn.Hosts, &tn.HostService{Ip: host, Name: "UncontactablePeer", Ports: []uint{16002}, LastSeen: time.Now()})
			}
			tn.ScanConfig()

			tn.ScanPublicInfo()
			tn.UniqueifyHosts()
		}
	}()
	go func() {

		for {
			log.Println("Sending hosts list to peers")
			tn.UpdatePeers()
			time.Sleep(time.Second * time.Duration(tn.Configuration.PeerUpdateInterval))
		}
	}()

	tn.Webserver(tn.Configuration.HttpPort, tn.Configuration.StartPagePort)

	for {
		time.Sleep(time.Second * 5)
	}
}

var netEntitiesLock sync.Mutex

func makeNetworkPcMenu(hosts []*tn.HostService) (*menu.Node, *menu.Node) {
	out := menu.MakeNodeLong("Network", []*menu.Node{}, "", "")
	global := menu.MakeNodeLong("Global Services", []*menu.Node{}, "", "")
	for _, host := range hosts {
		h := menu.MakeNodeLong(host.Ip+"/"+host.Name, []*menu.Node{}, host.Ip, "")
		for _, port := range host.Ports {
			protocol := "http"
			if port == 443 {
				protocol = "https"
			}
			h.SubNodes = append(h.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v(%v)", tn.PortMap()[int(port)], port), nil, fmt.Sprintf("%v://%v:%v/", protocol, host.Ip, port), ""))
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
	tn.Hosts = []*tn.HostService{}
	netmenus := menu.Node{Name: "Network", SubNodes: []*menu.Node{}}

	fmt.Printf("%+v, %v\n", m.SubNodes, m)

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

	go func() {

		if !noScan {
			tn.PortsToScan = append(tn.PortsToScan, tn.Configuration.HttpPort, tn.Configuration.StartPagePort)
			tn.ArpScan()
			tn.ScanC()
			tn.ScanConfig()

			tn.UniqueifyHosts()
			tn.ScanPublicInfo()
		} else {
			fmt.Println("Network scan disabled")
		}

		netmenu, globalmenu := makeNetworkPcMenu(tn.Hosts)

		netmenus.SubNodes = append(netmenus.SubNodes, netmenu)
		netmenus.SubNodes = append(netmenus.SubNodes, globalmenu)

	}()

}
