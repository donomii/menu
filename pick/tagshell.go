package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	//"strings"
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"

	//"sort"
	"net/rpc"

	"github.com/nsf/termbox-go"

	"sync"
)

var serverActive = false

var selection = 0
var itempos = 0
var cursorX = 11
var cursorY = 1
var selectPosX = 11
var selectPosY = 1
var focus = "input"
var inputPos = 0
var searchStr string
var debugStr = ""
var client *rpc.Client

var predictResults []string
var lines []string
var linesTr []ResultRecordTransmittable

var refreshMutex sync.Mutex

var LineCache map[string]string

func FetchLine(f string, lineNum int) (line string, lastLine int, err error) {
	key := fmt.Sprintf("%v%v", f, lineNum)
	if val, ok := LineCache[key]; ok {
		return val, -1, nil
	} else {
		r, _ := os.Open(f)
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			lastLine++
			if lastLine == lineNum {
				LineCache[key] = sc.Text()
				return sc.Text(), lastLine, sc.Err()
			}
		}
		LineCache[key] = line
		return line, lastLine, io.EOF
	}
}

//Contact server with search string
func search(searchTerm string, numResults int) []ResultRecordTransmittable {
	if statuses == nil {
		statuses = map[string]string{}
	}
	statuses["Status"] = "Searching"
	if searchTerm == "" {
		return linesTr
	}

	pr := MakeSearchPrint(RegSplit(strings.ToLower(searchTerm), SearchFragsRegex))

	var out []ResultRecordTransmittable
	for _, v := range linesTr {
		s := Score(pr, v)
		v.Score = fmt.Sprintf("%v", s)
		if s > 0 {
			out = append(out, v)
		}
		if len(out) > numResults {
			break
		}
	}

	statuses["Status"] = "Search complete"
	return out
}

var completeMatch = false

func isLinux() bool {
	return (runtime.GOOS == "linux")
}

func isDarwin() bool {
	return (runtime.GOOS == "darwin")
}

//Find the first space character to the left of the cursor
func searchLeft(aStr string, pos int) int {
	for i := pos; i > 0; i-- {
		if aStr[i-1] == ' ' {
			if pos != i {
				return i
			}
		}
	}
	return 0
}

//Find the first space character to the right of the cursor
func searchRight(aStr string, pos int) int {
	for i := pos; i < len(aStr)-1; i++ {
		if aStr[i+1] == ' ' {
			if pos != i {
				return i
			}
		}
	}
	return len(aStr) - 1
}

func extractWord(aLine string, pos int) string {
	start := searchLeft(aLine, pos)
	return aLine[start:pos]
}

//ForeGround colour
func foreGround() termbox.Attribute {
	return termbox.ColorBlack
}

//Background colour
func backGround() termbox.Attribute {
	return termbox.ColorWhite
}

//Display a string at XY
func putStr(x, y int, aStr string) {
	width, height := termbox.Size()
	if y >= height {
		return
	}
	for i, r := range aStr {
		if x+i >= width {
			return
		}
		termbox.SetCell(x+i, y, r, foreGround(), backGround())
	}
}

/*
//Redraw screen every 200 Milliseconds
func automaticRefreshTerm() {
	for i := 0; i < 1; i = 0 {
		refreshTerm()
		time.Sleep(time.Millisecond * 200)
		if !serverActive {
			statuses["Status"] = "Closed"
			return
		}
	}
}
*/
//Clean up and exit
func shutdown() {
	//Shut down resources so the display thread doesn't panic when the display driver goes away first
	//When we get a file persistence layer, it will go here
	statuses["Status"] = "Shutting down"

	serverActive = false
	os.Exit(0)

}

type LineIterator struct {
	reader *bufio.Reader
}

func NewLineIterator(rd io.Reader) *LineIterator {
	return &LineIterator{
		reader: bufio.NewReader(rd),
	}
}

func (ln *LineIterator) Next() ([]byte, error) {

	var bytes []byte
	for {
		line, isPrefix, err := ln.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, line...)
		if !isPrefix {
			break
		}
	}
	return bytes, nil
}

func slurpSTDIN() []string {
	arr := make([]string, 0)
	ln := NewLineIterator(os.Stdin)
	for {
		line, err := ln.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}
		arr = append(arr, string(line))
	}

	statuses["Input"] = fmt.Sprintf("%v lines", len(arr))
	return arr
}
func startPick() {
	log.Println("Starting pick")
	LineCache = map[string]string{}
	//flag.StringVar(&tagbrowser.ServerAddress, "server", tagbrowser.ServerAddress, fmt.Sprintf("Server IP and Port.  Default: %s", tagbrowser.ServerAddress))
	//flag.Parse()
	//terms := flag.Args()
	//if len(terms) < 1 {
	//	fmt.Println("Use: query.exe  < --completeMatch >  search terms")
	//}

	searchDir := "-"
	if len(flag.Args()) > 0 {
		searchDir = flag.Args()[0]
	}
	searchStr = ""

	predictResults = []string{}
	results = []ResultRecordTransmittable{}
	statuses = map[string]string{}

	//go automaticRefreshTerm()

	statuses["Input"] = "Reading"

	linesTr = []ResultRecordTransmittable{}
	if searchDir == "-" {
		log.Println("Reading lines from STDIN")
		lines := slurpSTDIN()
		for lineNum, l := range lines {
			r := ResultRecordTransmittable{"STDIN", fmt.Sprintf("%v", lineNum), MakeFingerprintFromData(l), l, "0"}
			//log.Printf("%+v\n", r)
			linesTr = append(linesTr, r)
		}
	} else {
		log.Println("Reading lines from ", searchDir)
		filepath.Walk(searchDir, func(fpath string, f os.FileInfo, err error) error {
			log.Printf("%v\n", fpath)
			path := fpath
			fileBytes, err := ioutil.ReadFile(path)

			lines := strings.Split(string(fileBytes), "\n")

			for lineNum, l := range lines {
				r := ResultRecordTransmittable{path, fmt.Sprintf("%v", lineNum), MakeFingerprintFromData(fmt.Sprintf("%v %v", l, path)), l, "0"}
				//log.Printf("%+v\n", r)
				linesTr = append(linesTr, r)
			}
			//statuses["File"] = path
			return nil
		})
	}
	statuses["Input"] = "Complete"
	log.Printf("Loaded lines %+v", linesTr)
	//results = search(searchStr, 50)
	//refreshTerm()

}
