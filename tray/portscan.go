package main

//https://gist.github.com/kotakanbe/d3059af990252ba89a82

import (
	"context"
	"fmt"

	"net"

	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func Ulimit() int64 {

	return 1000
}

func ScanPort(ip string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		//log.Println(err)
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		} else {
			//fmt.Println(ip, port, "closed")
		}
		return false
	}

	conn.Close()
	//fmt.Println(ip, port, "open")
	return true
}

func (ps *PortScanner) Start(f, l int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := f; port <= l; port++ {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func (ps *PortScanner) ScanList(f, l int, timeout time.Duration, ports []int) (out []int) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, port := range ports {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port int) {
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

type HostService struct {
	Ip    string
	Ports []int
}

func scanNetwork(cidr string, ports []int) (out []HostService) {
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
			openPorts := ps.ScanList(1, 9000, 1000*time.Millisecond, ports)
			if len(openPorts) > 0 {
				out = append(out, HostService{v, openPorts})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}

func scanIps(hosts []string, ports []int) (out []HostService) {
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
				out = append(out, HostService{v, openPorts})
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return out
}
