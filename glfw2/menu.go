package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/donomii/goof"
	"github.com/donomii/menu"
)

var Menu *menu.Node

func loadUserMenu() *menu.Node {
	var usermenu menu.Node
	exeDir := goof.ExecutablePath()
	b, _ := ioutil.ReadFile(exeDir + "/config/usermenu.json")
	json.Unmarshal(b, &usermenu)
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
