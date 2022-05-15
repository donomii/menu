//go:build go1.16
// +build go1.16

package traynetwork

import (
	//"embed"
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/donomii/menu"
)

var Info InfoStruct

type Config struct {
	HttpPort         uint
	StartPagePort    uint
	Name             string
	MaxUploadSize    uint
	Networks         []string
	ArpCheckInterval int
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

func contact(w http.ResponseWriter, req *http.Request) {
	//Get remote ip address from connection
	ip := req.RemoteAddr

	w.Write([]byte(fmt.Sprint(ip)))
}

func public_info(w http.ResponseWriter, req *http.Request) {
	out, _ := json.Marshal(Info)
	fmt.Fprintf(w, string(out))
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
	str := strings.Replace(string(base), "TEMPLATE", template(), -1)
	fmt.Fprintf(w, str)
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
		l = append(l, link{Label: item.Name, Url: strings.ReplaceAll(strings.ReplaceAll(item.Command, "\"", "'"), "\\", "/")})
	}
	return bookMarkMenu{Category: m.Name, Bookmarks: l}

}

func MakeUserMenu() *menu.Node {
	var usermenu menu.Node
	b, _ := ioutil.ReadFile("config/usermenu.json")
	json.Unmarshal(b, &usermenu)
	return &usermenu
}
func template() string {
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
	json.Unmarshal([]byte(data), &js)
	apps := menu.AppsMenu()
	//m := bookMarkMenu{Category: "User menu", Bookmarks: []link{link{Label: "user test", Url: "user link"}}}
	//newBookmark := link{Label: "test", Url: "test"}
	js = append(js, menu2jsmenu(apps))

	usermenu := MakeUserMenu()
	js = append(js, menu2jsmenu(usermenu))

	datab, err := json.Marshal(js)
	if err != nil {
		panic(err)
	}
	data = string(datab)

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
