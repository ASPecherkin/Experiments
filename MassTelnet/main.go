// +build linux darwin

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type cidr struct {
	Cidrs []string `json:"cidrs"`
}

func parseCidrs(path string) (c cidr, err error) {

	data, err := ioutil.ReadFile("./hosts.json")
	if err != nil {
		log.Fatalf("Couldn't read config file: %s  \n", err)
		return c, err
	}
	err = json.Unmarshal(data, &c)
	return c, nil
}

//Hosts return slice of hosts from given cidr address
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
	return ips[1 : len(ips)-1], nil
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

// Server stores ip and status
type Server struct {
	Adrress string
	Alive   bool
}

//IPStore storage for all ips generated from cidr addresses
type IPStore struct {
	sync.RWMutex
	Hosts []Server
}

func generateIPS(wg *sync.WaitGroup, cidrAddress string, out chan string) {
	defer wg.Done()
	hosts, err := Hosts(cidrAddress)
	if err != nil {
		log.Fatal(err)
	}
	for k := range hosts {
		out <- hosts[k]
	}
}

func storeIps(wg *sync.WaitGroup, in chan string, ips *IPStore) {
	defer ips.Unlock()
	ports := [4]string{":5228", ":5229", ":5230"}
	for ip := range in {
		ips.Lock()
		for _, v := range ports {
			ips.Hosts = append(ips.Hosts, Server{Adrress: ip + v})
		}
		ips.Unlock()
	}
}

func (s *Server) checkAddress(wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Printf("address %s \n", s.Adrress)
	_, err := net.DialTimeout("tcp", s.Adrress, 5*time.Second)
	if err != nil {
		s.Alive = false
	}
	s.Alive = true
}

func main() {
	var wg sync.WaitGroup
	hosts, err := parseCidrs("./hosts.json")
	if err != nil {
		log.Fatal(err)
	}
	pipe := make(chan string, 10)
	ipStore := IPStore{Hosts: make([]Server, 0, 100)}
	go storeIps(&wg, pipe, &ipStore)
	for k := range hosts.Cidrs {
		wg.Add(1)
		go generateIPS(&wg, hosts.Cidrs[k], pipe)
	}
	wg.Wait()
	fmt.Println(len(ipStore.Hosts))
	for k := range ipStore.Hosts {
		wg.Add(1)
		go ipStore.Hosts[k].checkAddress(&wg)
	}
	wg.Wait()
	fails, success := 0, 0
	for k := range ipStore.Hosts {
		if ipStore.Hosts[k].Alive {
			success++
		} else {
			fails++
		}
	}
	fmt.Printf("success = %d, fails = %d \n", success, fails)
}
