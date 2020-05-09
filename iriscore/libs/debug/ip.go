package debug

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func GetPublicIpByNet() (string, error) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.RemoteAddr().String(), ":")[0], nil

}

func GetInnerIpByNet() (string, error) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

/*
   10.0.0.0--10.255.255.255
   172.16.0.0--172.31.255.255
   192.168.0.0--192.168.255.255
*/
func GetInnerIpBySelf() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipaddr := addr.String()
		if strings.HasPrefix(ipaddr, "10.") || strings.HasPrefix(ipaddr, "192.") {
			iplist := strings.Split(ipaddr, "/")
			return iplist[0], nil
		}
		if ipaddr >= "172.16.0.0" && ipaddr <= "172.31.255.255" {
			iplist := strings.Split(ipaddr, "/")
			return iplist[0], nil
		}
	}

	return "", errors.New("no addr")
}

func ParseIpAddr(ipaddr string) (string, string, error) {
	ip, port, err := net.SplitHostPort(ipaddr)
	if err != nil {
		return "", "", err
	}
	if !IsIP(ip) {
		return "", "", fmt.Errorf("IP is invaliad %s", ip)
	}
	if !IsPort(port) {
		return "", "", fmt.Errorf("port is invaliad %s", port)
	}
	return ip, port, nil
}

func IpToLong(ip string) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long)
	return long
}

func LongToIP4(ipInt int64) string {

	// need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}
