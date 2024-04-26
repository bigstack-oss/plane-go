package nslookup

import (
	"net"

	planeLog "github.com/bigstack-oss/plane-go/pkg/base/log"
)

var (
	log, logf = planeLog.GetLoggers("nslookup-helper")
)

func ResolveIPs(domainName string) []string {
	ips, err := net.LookupIP(domainName)
	if err != nil {
		return nil
	}

	strIPs := []string{}
	for _, ip := range ips {
		strIPs = append(strIPs, ip.String())
	}

	return strIPs
}
