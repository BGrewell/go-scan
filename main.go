package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func work(workitems chan string, worksync *sync.WaitGroup, workersync *sync.WaitGroup, timeout int) {
	dialer := net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	}
	for addr := range workitems {
		c, err := dialer.Dial("tcp", addr)
		worksync.Done()
		if err != nil {
			continue
		}
		fmt.Printf("port %s open\n", strings.Split(addr, ":")[1])
		c.Close()
	}
	workersync.Done()
}

func main() {

	ip := flag.String("ip", "", "ip address of host")
	ports := flag.String("port", "1-1024", "port(s) or range to scan")
	threads := flag.Int("threads", 100, "number of worker threads")
	timeout := flag.Int("timeout", 1, "timeout in seconds to wait for a connection")
	flag.Parse()

	if *ip == "" {
		fmt.Println("required value 'ip' is missing")
		flag.Usage()
		os.Exit(1)
	}

	if *ports == "" {
		fmt.Println("required value 'ports' is missing")
		flag.Usage()
		os.Exit(1)
	}

	if *ports == "-" {
		*ports = "1-65535"
	}

	portsToScan := make([]int, 0)
	portGroups := strings.Split(*ports, ",")
	for _, group := range portGroups {
		portRanges := strings.Split(group, "-")
		if len(portRanges) == 2 {
			start, err := strconv.Atoi(portRanges[0])
			if err != nil {
				fmt.Printf("failed to parse port range %s\n", group)
				continue
			}
			end, err := strconv.Atoi(portRanges[1])
			if err != nil {
				fmt.Printf("failed to parse port range %s\n", group)
				continue
			}
			for i := start; i <= end; i++ {
				portsToScan = append(portsToScan, i)
			}
		} else {
			port, err := strconv.Atoi(portRanges[0])
			if err != nil {
				fmt.Printf("failed to parse port value %s\n", portRanges[0])
				continue
			}
			portsToScan = append(portsToScan, port)
		}
	}

	workitems := make(chan string, 65535)
	worksync := sync.WaitGroup{}
	for _, port := range portsToScan {
		addr := fmt.Sprintf("%s:%d", *ip, port)
		worksync.Add(1)
		workitems <- addr
	}

	workersync := sync.WaitGroup{}
	for i := 0; i < *threads; i++ {
		workersync.Add(1)
		go work(workitems, &worksync, &workersync, *timeout)
	}
	worksync.Wait()
	close(workitems)
	workersync.Wait()
	fmt.Printf("Done\n")
}
