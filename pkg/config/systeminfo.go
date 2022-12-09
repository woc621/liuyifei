package config

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

type SystemInfo struct {
	Os       string
	HostName string
	Ip       string
	Mac      string
}

func GetLocalIP(endpoints []string) (localip string) {
	for i, endpoint := range endpoints {
		conn, err := net.Dial("tcp", endpoint)
		defer conn.Close()
		if err == nil {
			localip, _, _ = net.SplitHostPort(conn.LocalAddr().String())
			return localip
		}
		if i == len(endpoints)-1 {
			localip = "127.0.0.1"
		}
	}
	return localip
}
func GetHWAddrByIp(localip string) (hwaddr string) {
	macs := make([]net.HardwareAddr, 0)
	inters, _ := net.Interfaces()
	for _, inter := range inters {
		addrs, _ := inter.Addrs()
		for _, addr := range addrs {
			ipaddr, _, _ := net.ParseCIDR(addr.String())
			if ipaddr.String() == localip {
				macs = append(macs, inter.HardwareAddr)
				for _, mac := range macs {
					hwaddr = fmt.Sprintf("%s", mac.String())

				}
				hwaddr = strings.ReplaceAll(hwaddr, ":", "")
			}
		}
	}
	return
}
func GetSystemInfo(endpoints []string) SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	localip := GetLocalIP(endpoints)
	hwaddr := GetHWAddrByIp(localip)
	os := runtime.GOOS
	fmt.Println(os)
	return SystemInfo{
		Os:       os,
		HostName: hostname,
		Ip:       localip,
		Mac:      hwaddr,
	}
}
