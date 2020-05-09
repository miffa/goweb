package debug

import (
	"net"
	"regexp"
	"strconv"
)

const (
	IP_REGEX_PATTERN     = `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	DOMAIN_REGEX_PATTERN = `[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?`
)

func IsIP(ip string) (b bool) {
	if m, _ := regexp.MatchString(IP_REGEX_PATTERN, ip); !m {
		//if m, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip); !m {
		return false
	}
	return true
}
func IsDomain(domaon string) (b bool) {
	if m, _ := regexp.MatchString(DOMAIN_REGEX_PATTERN, domaon); !m {
		return false
	}
	return true
}

func IsPort(port string) (b bool) {
	if mport, err := strconv.ParseInt(port, 10, 32); err != nil {
		return false
	} else if mport < 80 || mport > 65534 {
		return false
	} else {
		return true
	}
}
func Port(port string) int {
	if m, _ := regexp.MatchString("^[0-9]{1,3}", port); !m {
		return 0
	} else {
		if mport, err := strconv.ParseInt(port, 10, 32); err != nil {
			return 0
		} else if mport < 80 || mport > 65534 {
			return 0
		} else {
			return int(mport)
		}
	}
	return 0
}

func CheckIpAddr(ipaddr string) bool {
	ip, port, err := net.SplitHostPort(ipaddr)
	if err != nil {
		return false
	} else {
		return IsIP(ip) && IsPort(port)
	}
}
