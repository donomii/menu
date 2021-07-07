package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/donomii/goof"
	"github.com/mostlygeek/arp"
	"golang.org/x/sync/semaphore"
)

//https://gist.github.com/kotakanbe/d3059af990252ba89a82

type HostService struct {
	Ip       string
	Ports    []uint
	Services []Service
	Name     string
}

type HostServiceList []HostService

func (a HostServiceList) Len() int           { return len(a) }
func (a HostServiceList) Less(i, j int) bool { return a[i].Ip < a[j].Ip }
func (a HostServiceList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func Ulimit() int64 {

	return 1000
}

var hosts = []HostService{}
var PortsToScan = []uint{1, 5, 7, 9, 11, 13, 15, 17, 18, 19, 20, 21, 22, 23, 25, 53, 37, 42, 67, 68, 69, 70, 80, 88, 110, 119, 123, 137, 138, 139, 143, 177, 220, 445, 443, 514, 995, 989, 990, 1080, 3000, 3001, 3389, 8000, 8001, 8080, 8081, 8008, 16001, 16002}

func scanPorts() []uint {
	sp := PortsToScan
	sp = append(sp, Configuration.HttpPort, Configuration.StartPagePort)
	return sp
}
func ArpScan() {

	keys := []string{}
	for k, _ := range arp.Table() {
		keys = append(keys, k)
	}
	log.Printf("Scanning %v\n", keys)
	hosts = append(hosts, scanIps(keys, scanPorts())...)

}
func ScanC() {

	ips := goof.AllIps()

	classB := map[string]bool{}
	for _, ip := range ips {
		hosts = append(hosts, scanNetwork(ip+"/24", scanPorts())...)
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
		hosts = append(hosts, scanNetwork(network, scanPorts())...)
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
		url := fmt.Sprintf("http://%v:%v/public_info", v.Ip, Configuration.HttpPort)
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
					hosts[i].Name = s.Name
				}
			}
		}
	}

}

func ScanPort(ip string, port uint, timeout time.Duration) bool {
	if port == 0 {
		return false
	}
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		//log.Println(err)
		if strings.Contains(err.Error(), "open files") || strings.Contains(err.Error(), "requested address") {
			time.Sleep(timeout)
			fmt.Println(ip, ":", port, "retry :", err.Error())
			ScanPort(ip, port, timeout)
		} else {
			fmt.Println(ip, ":", port, "closed :", err.Error())
		}
		return false
	}

	conn.Close()
	//fmt.Println(ip, port, "open")
	return true
}

func (ps *PortScanner) Start(f, l uint, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := f; port <= l; port++ {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port uint) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func (ps *PortScanner) ScanList(f, l int, timeout time.Duration, ports []uint) (out []uint) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, port := range ports {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port uint) {
			defer ps.lock.Release(1)
			defer wg.Done()
			if ScanPort(ps.ip, port, timeout) {
				out = append(out, port)
			}
		}(port)
	}
	return out
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func scanNetwork(cidr string, ports []uint) (out []HostService) {
	var wg sync.WaitGroup
	hosts, _ := Hosts(cidr)
	for _, v := range hosts {
		wg.Add(1)
		//fmt.Println("Scanning", v)
		ps := &PortScanner{
			ip:   v,
			lock: semaphore.NewWeighted(Ulimit()),
		}
		go func(v string) {
			openPorts := ps.ScanList(1, 9000, 5000*time.Millisecond, ports)
			if len(openPorts) > 0 {
				out = append(out, HostService{v, openPorts, nil, ""})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}

func scanIps(hosts []string, ports []uint) (out []HostService) {
	var wg sync.WaitGroup

	for _, v := range hosts {
		wg.Add(1)
		//fmt.Println("Scanning", v)
		ps := &PortScanner{
			ip:   v,
			lock: semaphore.NewWeighted(Ulimit()),
		}
		go func(v string) {
			openPorts := ps.ScanList(1, 9000, 1000*time.Millisecond, ports)
			if len(openPorts) > 0 {
				out = append(out, HostService{v, openPorts, nil, ""})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}
