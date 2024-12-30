package utils

import (
	"net"
	"os"

	"github.com/bwmarrin/snowflake"
)

func CreateSnowFlakeNode() (*snowflake.Node, error) {
	// 生成机器ID
	var machine int64
	// 生成前半段by hostname
	hostName, _ := os.Hostname()
	var sum int64 = 0
	for _, b := range []byte(hostName) {
		sum += int64(b)
	}
	machine = (sum % 32) << 5
	// 生成后半段by ip
	ip := clientIP()
	sum = 0
	for _, b := range []byte(ip) {
		sum += int64(b)
	}
	machine = machine | (sum % 32)
	// 根据机器ID生成node
	node, err := snowflake.NewNode(machine)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func clientIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.To4()
			}
		}
	}
	return nil
}
