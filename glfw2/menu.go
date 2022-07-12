package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	

	"github.com/donomii/goof"
	menu ".."
)

var Menu *menu.Node

func LoadUserMenu() *menu.Node {
	var usermenu menu.Node
	var data []byte
	var configPath string =goof.HomePath(".umh/config/usermenu.json")
	if goof.Exists(configPath){
		fmt.Println("Loading menu from ",configPath)
		data, _ = ioutil.ReadFile(configPath)
	}else {
		exeDir := goof.ExecutablePath()
		configPath = exeDir + "/config/usermenu.json"
		data, _ = ioutil.ReadFile(configPath)
		fmt.Println("Loading menu from ",configPath)
	}
	
	err := json.Unmarshal(data, &usermenu)
	if err != nil {
	usermenu = menu.Node{Name: "Error in config file "+configPath, Command: "file://"+configPath}
	}
	return &usermenu
}

func renderMenuText(m *menu.Node) string {
	//AddSub(apps, appMen)
	out := ""
	for i, v := range m.SubNodes {

		out = out + fmt.Sprintf("%v ) %v <", i, v.Name)
		if len(v.SubNodes) > 0 {

			out = out + "<<<"

		} else {
		}

		out = out + "\n"

	}
	fmt.Println("Menu ", out)
	return out
}
