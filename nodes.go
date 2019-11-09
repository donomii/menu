// nodes.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/donomii/goof"
	"github.com/mattn/go-shellwords"
)

func (n *Node) String() string {
	return n.Name
}

func (n *Node) ToString() string {
	return n.Name
}

func makeStartNode() *Node {
	n := makeNodeShort("Command:", []*Node{})

	return n
}

func findNode(n *Node, name string) *Node {
	if n == nil {
		return n
	}
	for _, v := range n.SubNodes {
		if v.Name == name {
			return v
		}
	}
	return nil

}

func NodesToStringArray(ns []*Node) []string {
	var out []string
	for _, v := range ns {
		out = append(out, v.Name)

	}
	return out

}

func fileManagerMenu() *Node {
	return makeNodeShort("File Manager", []*Node{})
}
func appsMenu() *Node {
	node := makeNodeShort("Applications Menu",
		[]*Node{})
	addTextNodesFromStrStr(node, Apps())
	return node
}

var appCache [][]string

func Apps() [][]string {

	out := [][]string{}
	switch runtime.GOOS {
	//case "linux":
	case "windows":
		for _, progDir := range []string{"ProgramData", "AppData"} {
			appPath := os.Getenv(progDir) + "\\Microsoft\\Windows\\Start Menu\\Programs\\"
			log.Println("Loading apps from", appPath)
			lines := goof.LslR(appPath)

			for _, v := range lines {
				if strings.HasSuffix(v, ".lnk") {
					name := strings.TrimSuffix(v, ".lnk")
					name = strings.TrimPrefix(name, appPath)
					command := fmt.Sprintf("!%v", v)
					out = append(out, []string{name, command})
				}
			}
		}
		log.Println(out)

	case "darwin":
		lines := goof.Ls("/Applications")

		for _, v := range lines {
			name := strings.TrimSuffix(v, ".app")
			command := fmt.Sprintf("!open \"/Applications/%v\"", v)
			out = append(out, []string{name, command})
		}

	case "linux":
		src := goof.Command("find", []string{"/usr/share/applications", "~/.local/share/applications", "-name", "*.desktop"})
		lines := strings.Split(src, "\n")
		out := [][]string{}
		for _, v := range lines {
			content, _ := ioutil.ReadFile(v)
			contents := strings.Split(string(content), "\n")
			matches := goof.ListGrep("Exec=", contents)
			if len(matches) > 0 {
				bits := strings.Split(matches[0], "=") //FIXME
				exeString := bits[1]
				displayName := path.Base(v)
				//fmt.Println(displayName,"|",exeString)
				out = append(out, []string{displayName, " " + exeString})
			}
		}
	default:
		log.Println("unsupported platform when trying to get applications")
	}

	if appCache == nil {
		appCache = out
		return out
	} else {
		return appCache
	}

}

func controlMenu() *Node {
	node := makeNodeShort("System controls", []*Node{})
	addTextNodesFromStrStr(node,
		[][]string{
			[]string{"pmset sleepnow"},
		})
	return node
}

func historyMenu() *Node {
	return addHistoryNodes()
}

func addHistoryNodes() *Node {
	src := goof.Command("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := makeNodeShort("Previous command lines", []*Node{})
	for _, l := range lines {
		currentNode := startNode
		/*
				var s scanner.Scanner
				s.Init(strings.NewReader(l))
				s.Filename = "example"
				for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			        text := s.TokenText()
					fmt.Printf("%s: %s\n", s.Position, text)
			        if findNode(currentNode, text) == nil {
			            newNode := Node{text, []*Node{}}
			            currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
			            currentNode = &newNode
			        } else {
			            currentNode = findNode(currentNode, text)
			        }
		*/
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}

		}
	}
	return startNode
}

func addTextNodesFromString(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	return addTextNodesFromStringList(startNode, lines)
}

func appendNewNodeShort(text string, aNode *Node) *Node {
	newNode := makeNodeShort(text, []*Node{})
	aNode.SubNodes = append(aNode.SubNodes, newNode)
	return aNode
}

func addTextNodesFromStringList(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}
		}
	}

	return startNode

}

func addTextNodesFromCommands(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		appendNewNodeShort(l, startNode)
	}

	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], ""}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	return startNode

}

func addTextNodesFromStrStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], l[2]}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	return startNode

}

func dumpTree(n *Node, indent int) {
	fmt.Printf("%*s%s\n", indent, "", n.Name)
	for _, v := range n.SubNodes {
		dumpTree(v, indent+1)
	}

}
