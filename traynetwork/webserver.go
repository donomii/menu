package traynetwork

import (
	//"embed"
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"net/url"
	"strings"
	"time"

	menu ".."
)

var Info InfoStruct

type Config struct {
	HttpPort           uint
	StartPagePort      uint
	Name               string
	MaxUploadSize      uint
	Networks           []string
	KnownPeers         []string
	ArpCheckInterval   int
	PeerUpdateInterval int
}

type Service struct {
	Name        string
	Ip          string
	Port        int
	Protocol    string
	Description string
	Global      bool
	Path        string
}

var Configuration Config

type InfoStruct struct {
	Name     string
	Services []Service
}

func LoadConfig() {
	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Configuration)
	if err != nil {
		panic(err)
	}
	for _, host := range Configuration.KnownPeers {
		Hosts = append(Hosts, &HostService{Ip: host, Name: host, Ports: []uint{16002}, LastSeen: time.Now()})
		log.Printf("Added known peer %v\n", host)
	}
	fmt.Printf("Loaded config: %+v\n", Configuration)
}

//go:embed webfiles/*
var webapp embed.FS

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, Configuration.Name)
}

func landingPage(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(landingTemplate()))
}

func Hosts2Json() []byte {
	out, _ := json.Marshal(Hosts)
	return out
}
func contact(w http.ResponseWriter, req *http.Request) {

	//Get remote ip address from connection
	ip := req.RemoteAddr
	//Read a json struct from the request body
	var data []*HostService
	req.ParseForm()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Failed to read request body", err)
		panic(err)
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Failed to unmarshal request body", err, string(body))
		panic(err)
	}

	//Add the remote ip to the list of hosts
	log.Println("Received contact from:", ip)
	log.Printf("Received hosts list: %+v\n", data)

	Hosts = append(Hosts, data...)
	UniqueifyHosts()

	w.Write(Hosts2Json())
}

func public_info(w http.ResponseWriter, req *http.Request) {
	out, _ := json.Marshal(Info)
	fmt.Fprintf(w, string(out))
}

func UpdatePeers() {

	for _, host := range Configuration.KnownPeers {
		Hosts = append(Hosts, &HostService{Ip: host, Name: host, Ports: []uint{16002}})
		log.Printf("Added known peer %v\n", host)
	}
	for _, host := range Hosts {
		//Post the hosts list to the host
		data, _ := json.Marshal(Hosts)
		log.Printf("Sending hosts list to http://%v:%v/contact", host.Ip, Configuration.HttpPort)
		resp, err := http.Post(fmt.Sprintf("http://%v:%v/contact", host.Ip, Configuration.HttpPort), "application/json", ioutil.NopCloser(strings.NewReader(string(data))))
		if err != nil {
			log.Println("Failed to send hosts list to", host.Ip, "err:", err)
		} else {

			//Read entire body from response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Failed to read response body from", host.Ip, "err:", err)
			} else {
				log.Println("Received response from", host.Ip, ":", string(body))
				var h []*HostService
				err = json.Unmarshal(body, &h)
				if err != nil {
					log.Println("Failed to unmarshal response body from", host.Ip, "err:", err)
				} else {
					log.Printf("Received hosts list from%v: %+v", host.Ip, h)
					Hosts = append(Hosts, h...)
					UniqueifyHosts()
				}

			}

			resp.Body.Close()
		}

	}
	ScanPublicInfo()
}

func Webserver(apiport, startpageport uint) {

	server1 := http.NewServeMux()
	server1.HandleFunc("/upload", uploadHandler)
	server1.HandleFunc("/hello", hello)
	server1.HandleFunc("/public_info", public_info)
	server1.HandleFunc("/contact", contact)

	log.Println("Server started on: :", apiport)
	go http.ListenAndServe(fmt.Sprintf(":%v", apiport), server1)

	server2 := http.NewServeMux()
	fs := http.FileServer(http.FS(webapp))
	server2.Handle("/", fs)

	server2.HandleFunc("/webfiles/js/index.js", fillTemplate)

	log.Println("Server started on: 0.0.0.0:", startpageport)
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", startpageport), server2)
}

func landingTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>UMH</title>
  </head>
  <body>
<h1>UMH Menu</h1>
Web service interface to UMH menu.

<a href="/upload">Upload</a> a file, or view the <a href="/public_info">details</a> for this computer
  </body>
</html>`

}

func fillTemplate(w http.ResponseWriter, req *http.Request) {
	base, err := webapp.ReadFile("webfiles/js/index.js")
	if err != nil {
		panic(err)
	}
	tem := template()
	tem = strings.ReplaceAll(tem, "\"", "\\\"")
	str := strings.Replace(string(base), "TEMPLATE", tem, -1)
	fmt.Fprint(w, str)
}

type link struct {
	Label string `json:"label"`
	Url   string `json:"url"`
}

type bookMarkMenu struct {
	Category  string `json:"category"`
	Bookmarks []link `json:"bookmarks"`
}

func menu2jsmenu(m *menu.Node) bookMarkMenu {
	l := []link{}

	for _, item := range m.SubNodes {

		urlstr := strings.ReplaceAll(strings.ReplaceAll(item.Command, "\"", "'"), "\\", "/")
		if !strings.HasPrefix(item.Command, "http") {

			urlstr = fmt.Sprintf("http://localhost:%v/command/%v", Configuration.HttpPort, urlstr)
		}

		l = append(l, link{
			Label: item.Name,
			Url:   item.Command,
		},
		)
	}

	log.Printf("JSON for webpage: %v\n", l)
	return bookMarkMenu{Category: m.Name, Bookmarks: l}

}

func MakeUserMenu() *menu.Node {
	var usermenu menu.Node
	b, _ := ioutil.ReadFile("config/usermenu.json")
	json.Unmarshal(b, &usermenu)
	return &usermenu
}

func MakeNetworkPcMenu(hosts []*HostService) (*menu.Node, *menu.Node) {
	log.Printf("Hosts: %v\n", hosts)
	out := menu.MakeNodeLong("Network", []*menu.Node{}, "", "")
	global := menu.MakeNodeLong("Global Services", []*menu.Node{}, "", "")
	for _, host := range hosts {
		//If we have seen the host in the last 10 minutes, add it to the menu
		if host.LastSeen.Add(10 * time.Minute).After(time.Now()) {

			NodeName := host.Name
			h := menu.MakeNodeLong(host.Ip+"/"+host.Name, []*menu.Node{}, "http://"+host.Ip, "")
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
					out.SubNodes = append(out.SubNodes, menu.MakeNodeLong(fmt.Sprintf("%v %v", NodeName, s.Name), nil, fmt.Sprintf("%v://%v:%v%v", protocol, ip, s.Port, s.Path), ""))
				}
			}
			out.SubNodes = append(out.SubNodes, h)

		}
	}
	return out, global
}

func template() string {
	netmenu, _ := MakeNetworkPcMenu(Hosts)
	data := `[
        {
            "category": "Social Media",
            "bookmarks": [
                { "label": "Facebook",              "url": "https://www.facebook.com" },
                { "label": "Messenger",             "url": "https://www.messenger.com" },
                { "label": "Instagram",             "url": "https://www.instagram.com" },
                { "label": "Reddit",                "url": "https://www.reddit.com" },
                { "label": "Twitter",               "url": "https://www.twitter.com" }
            ]
        }
    ]`

	var js []bookMarkMenu
	//json.Unmarshal([]byte(data), &js)
	apps := menu.AppsMenu()
	//m := bookMarkMenu{Category: "User menu", Bookmarks: []link{link{Label: "user test", Url: "user link"}}}
	//newBookmark := link{Label: "test", Url: "test"}
	js = append(js, menu2jsmenu(netmenu))
	js = append(js, menu2jsmenu(apps))

	usermenu := MakeUserMenu()
	js = append(js, menu2jsmenu(usermenu))
	log.Printf("vals for webpage: %v\n", js)

	datab, err := json.Marshal(js)
	if err != nil {
		panic(err)
	}
	data = string(datab)
	log.Printf("\n\nJSON for webpage: %v\n", data)

	jsText := `{
    "bookmarks": TEMPLATE,

    "bookmarkOptions": {
        "alwaysOpenInNewTab": true,
        "useFaviconKit": false
    },

    "sidebar": {
        "idleOpacity": 0.06
    },

    "voiceReg": {
        "enabled": true,
        "language": "en-US"
    },
    
    "glass": {
        "background": "rgba(47, 43, 48, 0.568)",
        "backgroundHover": "rgba(47, 43, 48, 0.568)",
        "editorBackground": "rgba(0,0,0, 0.868)",
        "blur": 12
    },

    "background": {
        "url": "https://wallpaperaccess.com/full/7285.jpg",
        "snow": {
            "enabled": false,
            "count": 200
        },
        "mist": {
            "enabled": false,
            "opacity": 5
        },
        "css": "filter: blur(0px) saturate(150%); transform: scale(1.1); opacity: 1"
    }
}`
	str := strings.Replace(string(jsText), "TEMPLATE", data, -1)
	return str

}
