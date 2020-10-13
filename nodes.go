// nodes.go
package menu

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github.com/donomii/goof"
	"github.com/mattn/go-shellwords"
	"github.com/schollz/closestmatch"
)

type Node struct {
	Name     string
	SubNodes []*Node
	Command  string
	Data     string
	Function func()
}

func MakeNodeShort(name string, subNodes []*Node) *Node {
	if subNodes == nil {
		subNodes = []*Node{}
	}
	return &Node{name, subNodes, name, "", nil}
}

func MakeNodeLong(name string, subNodes []*Node, command, data string) *Node {
	if subNodes == nil {
		subNodes = []*Node{}
	}
	return &Node{name, subNodes, name, data, nil}
}

func (n *Node) String() string {
	return n.Name
}

func (n *Node) ToString() string {
	return n.Name
}

func MakeStartNode() *Node {
	n := MakeNodeShort("Command:", []*Node{})

	return n
}

func FindNode(n *Node, name string) *Node {
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

func AppendNode(menu, item *Node) {
	menu.SubNodes = append(menu.SubNodes, item)
}

func fileManagerMenu() *Node {
	return MakeNodeShort("File Manager", []*Node{})
}

var appCache [][]string

func AppsMenu() *Node {
	node := MakeNodeShort("Applications Menu",
		[]*Node{})
	AddTextNodesFromStrStr(node, Apps())
	return node
}

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

		lines = goof.Ls("/Applications/Utilities")

		for _, v := range lines {
			name := strings.TrimSuffix(v, ".app")
			command := fmt.Sprintf("!open \"/Applications/Utilities/%v\"", v)
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

func ControlMenu() *Node {
	node := MakeNodeShort("System controls", []*Node{})
	AddTextNodesFromCommands(node, []string{"pmset sleepnow"})
	return node
}

func HistoryMenu() *Node {
	return addHistoryNodes()
}

func addHistoryNodes() *Node {
	src := goof.Command("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := MakeNodeShort("Previous command lines", []*Node{})
	for _, l := range lines {
		currentNode := startNode
		/*
				var s scanner.Scanner
				s.Init(strings.NewReader(l))
				s.Filename = "example"
				for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			        text := s.TokenText()
					fmt.Printf("%s: %s\n", s.Position, text)
			        if FindNode(currentNode, text) == nil {
			            newNode := Node{text, []*Node{}}
			            currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
			            currentNode = &newNode
			        } else {
			            currentNode = FindNode(currentNode, text)
			        }
		*/
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if FindNode(currentNode, text) == nil {
				newNode := MakeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = FindNode(currentNode, text)
			}

		}
	}
	return startNode
}

func AddTextNodesFromString(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	return AddTextNodesFromStringList(startNode, lines)
}

func appendNewNodeShort(text string, aNode *Node) *Node {
	newNode := MakeNodeShort(text, []*Node{})
	aNode.SubNodes = append(aNode.SubNodes, newNode)
	return aNode
}

func AddTextNodesFromStringList(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if FindNode(currentNode, text) == nil {
				newNode := MakeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = FindNode(currentNode, text)
			}
		}
	}

	return startNode

}

func AddTextNodesFromCommands(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		appendNewNodeShort(l, startNode)
	}

	dumpTree(startNode, 0)
	return startNode

}

func AddTextNodesFromStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := *MakeNodeLong(l[0], []*Node{}, l[1], "")
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	return startNode

}

func AddTextNodesFromStrStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := *MakeNodeLong(l[0], []*Node{}, l[1], l[2])
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

var recallCache [][]string

func Predict(newString []byte) []string {
	news := string(newString)
	//log.Println("Processing ", news)
	if appCache == nil {
		appCache = Apps()
	}
	if recallCache == nil {
		recallCache = Recall()
	}
	var names []string
	for _, details := range appCache {
		names = append(names, details[0])
	}
	for _, details := range recallCache {
		names = append(names, details[0])
	}
	wordsToTest := names

	//log.Println("Words to test: ", wordsToTest)

	// Choose a set of bag sizes, more is more accurate but slower
	bagSizes := []int{2}

	// Create a closestmatch object
	cm := closestmatch.New(wordsToTest, bagSizes)
	return cm.ClosestN(news, 5)
}

func loadEnsureRecallFile(recallFile string) []byte {
	var raw []byte
	if goof.Exists(recallFile) {

		raw, _ = ioutil.ReadFile(recallFile)
	} else {
		log.Println("Writing default configuration file to", recallFile)
		raw = []byte(fmt.Sprintf("Recall Config File Location | %v\nReddit | http://reddit.com\nMy password | AbCdEfG", recallFile))
		ioutil.WriteFile(recallFile, raw, 0600)
	}
	return raw
}

func RecallFilePath() string {
	return goof.ConfigFilePath(".menu.recall.txt")
}

func Recall() [][]string {
	recallFile := RecallFilePath()
	log.Println("Reading default configuration file from", recallFile)

	raw := loadEnsureRecallFile(recallFile)
	lines := strings.Split(string(raw), "\n")
	out := [][]string{}
	for _, v := range lines {
		//name := strings.TrimSuffix(v, ".app")
		name := v
		command := "recall"
		out = append(out, []string{name, command})
	}
	return out
}

func Activate(value string) bool {
	result := ""
	log.Println("selected for activation:", value)
	appCache := Apps()

	for i, v := range appCache {
		cmp := strings.Compare(value, v[0])

		if cmp == 0 {

			cmd := appCache[i][1][1:]

			switch runtime.GOOS {
			case "linux":
				log.Println("Starting ", cmd)
				result = goof.Command("/bin/sh", []string{"-c", cmd})
				result = result + goof.Command("cmd", []string{"/c", cmd})
				return true
			case "windows":
				cmdArray := []string{"/c", cmd}
				log.Println("Starting cmd", cmdArray)
				go goof.Command("c:\\Windows\\System32\\cmd.exe", cmdArray)
				time.Sleep(5 * time.Second) //FIXME use cmd.Exec or w/e to start program then exit
				return true
			case "darwin":
				result = result + goof.Command("/bin/sh", []string{"-c", cmd})
				return true
			default:
				log.Println("unsupported platform when trying to run application")
			}

		}
		if recallCache == nil {
			recallCache = Recall()
		}
		for _, v := range recallCache {
			name := v[0]
			//log.Println("Searching for", value, name)
			cmp := strings.Compare(value, name)
			if cmp == 0 {
				//log.Println("Found", value, v[1])
				if v[1] == "recall" {

					//Copy to clipboard
					bits := strings.SplitN(name, " | ", 2)
					data := bits[1]
					if strings.HasPrefix(data, "http") {
						url := data
						log.Println("Opening ", data, "in browser")
						var err error
						switch runtime.GOOS {
						case "linux":
							goof.QC([]string{"xdg-open", url})
						case "windows":
							goof.QC([]string{"rundll32", "url.dll,FileProtocolHandler"})
						case "darwin":

							goof.QC([]string{"open", url})
						default:
							err = fmt.Errorf("unsupported platform")
						}
						if err != nil {
							log.Println(err)
						}
						return true
					}
					if strings.HasPrefix(data, "file") {
						fmt.Println("Opening for edit: ", data)

						//goof.QC([]string{"open", recallFile})
						go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", data})
						go goof.Command("/usr/bin/open", []string{data})
					}

					log.Println("Copying ", data, "to clipboard")
					if err := clipboard.WriteAll(data); err != nil {
						panic(err)
					}

					return true
				}

			}
		}
	}
	return false

}
