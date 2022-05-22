package traynetwork

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
	LastSeen time.Time
}

type HostServiceList []*HostService

func (a HostServiceList) Len() int           { return len(a) }
func (a HostServiceList) Less(i, j int) bool { return a[i].Ip < a[j].Ip }
func (a HostServiceList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func Ulimit() int64 {

	return 1
}

var GlobalScanSemaphore = semaphore.NewWeighted(5)

var Hosts = []*HostService{}

//var PortsToScan = []uint{1, 20, 21, 22, 23, 25, 80, 443, 8000, 8001, 8080, 8081, 8008, 9000, 16001, 16002}
var PortsToScan = []uint{22, 80, 443}

func scanPorts() []uint {
	sp := PortsToScan
	return sp
}

var AppendHostsLock = &sync.Mutex{}

func AppendHosts(hs []*HostService) {
	AppendHostsLock.Lock()
	defer AppendHostsLock.Unlock()
	Hosts = append(Hosts, hs...)
}
func ArpScan() {

	keys := []string{}
	for k, _ := range arp.Table() {
		keys = append(keys, k)
	}
	log.Printf("Scanning %v\n", keys)
	Hosts = append(Hosts, scanIps(keys, scanPorts())...)

}
func ScanC() {

	ips := goof.AllIps()

	classB := map[string]bool{}
	for _, ip := range ips {
		Hosts = append(Hosts, scanNetwork(ip+"/24", scanPorts())...)
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
		Hosts = append(Hosts, scanNetwork(network, scanPorts())...)
	}

}

func UniqueifyHosts() {
	temp := map[string]*HostService{}
	for _, v := range Hosts {
		if _, ok := temp[v.Ip]; !ok {
			temp[v.Ip] = v
		} else {
			if v.LastSeen.After(temp[v.Ip].LastSeen) {
				temp[v.Ip] = v
			}
		}
	}

	out := HostServiceList{}
	for _, v := range temp {
		out = append(out, v)
	}
	sort.Sort(out)
	Hosts = out
}
func ScanPublicInfo() {

	for _, v := range Hosts {
		url := fmt.Sprintf("http://%v:%v/public_info", v.Ip, Configuration.HttpPort)
		fmt.Println("Public info url:", url)
		resp, err := http.Get(url)
		if err == nil {
			log.Println("Got response")
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				log.Println("Got body")
				var s InfoStruct
				err := json.Unmarshal(body, &s)
				if err == nil {
					fmt.Printf("Unmarshalled body %v", s)
					v.Services = s.Services
					v.Name = s.Name
					v.LastSeen = time.Now()
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
	//AddNetworkNode(ip, "http", port)
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

func CidrHosts(cidr string) ([]string, error) {
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

func scanNetwork(cidr string, ports []uint) (out []*HostService) {
	var wg sync.WaitGroup
	hosts, _ := CidrHosts(cidr)
	for _, v := range hosts {
		wg.Add(1)
		//fmt.Println("Scanning", v)
		ps := &PortScanner{
			ip:   v,
			lock: GlobalScanSemaphore,
		}
		go func(v string) {
			openPorts := ps.ScanList(1, 9000, 5000*time.Millisecond, ports)
			if len(openPorts) > 0 {
				out = append(out, &HostService{v, openPorts, nil, "", time.Now()})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}

func scanIps(hosts []string, ports []uint) (out []*HostService) {
	var wg sync.WaitGroup

	for _, v := range hosts {
		wg.Add(1)
		//fmt.Println("Scanning", v)
		ps := &PortScanner{
			ip:   v,
			lock: GlobalScanSemaphore,
		}
		go func(v string) {
			openPorts := ps.ScanList(1, 9000, 3000*time.Millisecond, ports)
			if len(openPorts) > 0 {
				out = append(out, &HostService{v, openPorts, nil, "", time.Now()})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}
