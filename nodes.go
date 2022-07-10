// nodes.go
package menu

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/atotto/clipboard"
	"github.com/emersion/go-autostart"

	"github.com/donomii/goof"
	"github.com/mattn/go-shellwords"
	"github.com/schollz/closestmatch"
)

type Node struct {
	Name     string
	SubNodes []*Node
	Command  string
	Data     string
	Function func() `json:"-"`
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
	return &Node{name, subNodes, command, data, nil}
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
	node := MakeNodeShort("Applications Menu", []*Node{})
	AddTextNodesFromStrStr(node, Apps())
	return node
}

func TieredAppsMenu() *Node {
	node := MakeNodeShort("Applications Menu", []*Node{})
	cwd, _ := os.Getwd()
	switch runtime.GOOS {
	//case "linux":
	case "windows":
		for _, progDir := range []string{"ProgramData", "AppData"} {
			appPath := os.Getenv(progDir) + "\\Microsoft\\Windows\\Start Menu\\Programs\\"
			os.Chdir(appPath)
			Dir2Menu(node)
		}
	case "darwin":

		for _, progDir := range []string{"~/Applications", "/Applications"} {
			os.Chdir(progDir)
			Dir2Menu(node)
		}
	}
	os.Chdir(cwd)
	return node
}

func Dir2Menu(parent *Node) *Node {

	lines := goof.Ls(".")

	for _, v := range lines {
		if goof.IsDir(v) {
			node := MakeNodeShort(v, []*Node{})
			cwd, _ := os.Getwd()
			os.Chdir(v)
			node = Dir2Menu(node)
			AppendNode(parent, node)
			os.Chdir(cwd)
		} else {
			switch runtime.GOOS {
			//case "linux":
			case "windows":
				if strings.HasSuffix(v, ".lnk") {
					cwd, _ := os.Getwd()

					command := fmt.Sprintf("shell://"+cwd+"\\%v", v)
					name := strings.TrimSuffix(v, ".lnk")
					//name = strings.TrimPrefix(name, appPath)
					name = filepath.Base(name)
					node := Node{Name: name, Command: command}
					AppendNode(parent, &node)
				}
			case "darwin":
				if strings.HasSuffix(v, ".lnk") {
					cwd, _ := os.Getwd()
					name := strings.TrimSuffix(v, ".app")
					command := fmt.Sprintf("shell://open \"%v/%v\"", cwd, v)
					node := Node{Name: name, Command: command}
					AppendNode(parent, &node)
				}
			}
		}
	}
	return parent
}

//Get a list of all (major gui) apps installed on the system
func Apps() [][]string {

	out := [][]string{}
	switch runtime.GOOS {
	//case "linux":
	case "windows":
		for _, progDir := range []string{"ProgramData", "AppData"} {
			appPath := os.Getenv(progDir) + "\\Microsoft\\Windows\\Start Menu\\Programs\\"
			//log.Println("Loading apps from", appPath)
			lines := goof.LslR(appPath)

			for _, v := range lines {
				if strings.HasSuffix(v, ".lnk") {
					command := fmt.Sprintf("shell://%v", v)
					name := strings.TrimSuffix(v, ".lnk")
					//name = strings.TrimPrefix(name, appPath)
					name = filepath.Base(name)

					out = append(out, []string{name, command})
				}
			}
		}
		//log.Println(out)

	case "darwin":
		lines := goof.Ls("/Applications")

		for _, v := range lines {
			name := strings.TrimSuffix(v, ".app")
			command := fmt.Sprintf("shell://open \"/Applications/%v\"", v)
			out = append(out, []string{name, command})
		}

		lines = goof.Ls("/Applications/Utilities")

		for _, v := range lines {
			name := strings.TrimSuffix(v, ".app")
			command := fmt.Sprintf("shell://open \"/Applications/Utilities/%v\"", v)
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
	//log.Printf("out %v", out)

	//FIXME
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

var RecallCache [][]string

func Predict(newString []byte) ([]string, []string) {
	news := string(newString)
	//log.Println("Processing ", news)
	if appCache == nil {
		appCache = Apps()
	}
	if RecallCache == nil {
		RecallCache = Recall()
	}
	allEntries := [][]string{}
	var names []string

	for _, details := range appCache {
		names = append(names, details[0])
		allEntries = append(allEntries, details)
	}

	for _, details := range RecallCache {
		names = append(names, details[0])
		allEntries = append(allEntries, details)
	}
	wordsToTest := names

	log.Printf("Words to test: %#v\n", wordsToTest)

	// Choose a set of bag sizes, more is more accurate but slower
	bagSizes := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	// Create a closestmatch object
	cm := closestmatch.New(wordsToTest, bagSizes)
	log.Printf("Finding matches for: %#v\n", news)
	predictions := cm.ClosestN(news, 5)

	log.Printf("Predicted: %#v\n", predictions)

	var out, actions []string
	for _, pred := range predictions {
		for i, v := range allEntries {
			cmp := strings.Compare(pred, v[0])

			if cmp == 0 {

				cmd := allEntries[i][1]
				out = append(out, pred)
				actions = append(actions, cmd)

			}
		}
	}
	return out, actions
}

func loadEnsureRecallFile(recallFile string) []byte {
	var raw []byte
	var err error
	if goof.Exists(recallFile) {
		log.Println("File exists, reading", recallFile)
		raw, err = ioutil.ReadFile(recallFile)
		if err != nil {
			panic(err)
		}
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
	log.Printf("Recall file contents: %v\n", string(raw))
	lines := strings.Split(string(raw), "\n")
	out := [][]string{}
	for _, v := range lines {
		bits := strings.Split(v, "|")
		if len(bits) < 2 {
			//Vi and friends add a blank line to the end of the file
			//fmt.Printf("Failed to parse line: %v\n", v)
			continue
		}
		name := bits[0]
		name = strings.TrimSpace(name)

		command := bits[1]
		command = strings.TrimSpace(command)
		entry := []string{name, command}
		log.Printf("Entry: %v\n", entry)
		out = append(out, entry)
	}
	return out
}

type cmdSubs struct {
	AppDir    string
	ConfigDir string
	Cwd       string
	Command   string
}

func Activate(value string) bool {
	result := ""
	log.Println("selected for activation:", value)

	subs := cmdSubs{goof.ExecutablePath(), goof.HomePath(".umh/"), goof.Cwd(), value}
	tmpl, err := template.New("test").Parse(value)
	if err != nil {
		panic(err)
	}
	var s string
	buf := bytes.NewBufferString(s)
	err = tmpl.Execute(buf, subs)
	value = buf.String()
	fmt.Printf("Resolved template:%v\n", value)

	if strings.HasPrefix(value, "internal://") {
		cmd := strings.TrimPrefix(value, "internal://")
		log.Println("Executing internal", cmd)
		exePath, err := os.Executable()
		if err != nil {
			cwd, _ := os.Getwd()
			exePath = cwd + "/" + "tray.exe"
		}

		switch cmd {
		case "EditRecallFile":
			log.Println("Editing recall file")
			//Get user directory
			homedir := goof.HomeDirectory()
			filePath := homedir + "/.menu.recall.txt"
			value = "file://" + filePath
		case "RunAtStartup":
			app := &autostart.App{
				Name:        "umhtray",
				DisplayName: "UMH Tray",
				Exec:        []string{exePath},
			}
			app.Enable()
		case "Exit":
			os.Exit(0)
		case "Reload":
			RecallCache = nil
		default:
			log.Println("unsupported command when trying to run", value)

		}
	}

	if strings.HasPrefix(value, "exec://") {
		cmd := strings.TrimPrefix(value, "exec://")
		log.Println("Executing", cmd)
		cmdline, err := shellwords.Parse(cmd)
		cmdline[0], err = filepath.Abs(cmdline[0])
		log.Println(err)
		go goof.QC(cmdline)
	}

	if strings.HasPrefix(value, "shell://") {
		cmd := strings.TrimPrefix(value, "shell://")

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
			//time.Sleep(5 * time.Second) //FIXME use cmd.Exec or w/e to start program then exit
			return true
		case "darwin":
			result = result + goof.Command("/bin/sh", []string{"-c", cmd})
			return true
		default:
			log.Println("unsupported platform when trying to run application")
		}
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		url := value
		log.Println("Opening ", url, "in browser")
		var err error
		switch runtime.GOOS {
		case "linux":
			goof.QC([]string{"xdg-open", url})
		case "windows":
			go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", url})
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

	if strings.HasPrefix(value, "file://") {
		data := strings.TrimPrefix(value, "file://")
		fmt.Println("Opening for edit: ", data)

		//goof.QC([]string{"open", recallFile})
		go goof.Command("c:\\Windows\\System32\\cmd.exe", []string{"/c", "start", data})
		go goof.QC([]string{"rundll32", "url.dll,FileProtocolHandler"})
		go goof.Command("/usr/bin/open", []string{data})
		return true
	}

	if strings.HasPrefix(value, "clipboard://") {
		data := strings.TrimPrefix(value, "clipboard://")
		log.Println("Copying ", data, "to clipboard")
		if err := clipboard.WriteAll(data); err != nil {
			panic(err)
		}
		return true
	}

	return false
}
