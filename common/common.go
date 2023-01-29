// Package common 公用方法
package common

import (
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

const CanNotGetIp = "0x3f3f"

// GetOutBoundIP 获取本机IP地址
func GetOutBoundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		defer conn.Close()
		return CanNotGetIp, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(localAddr.IP.String(), ":")[0]
	logrus.Infof("get localhost IP success! ip = %s", ip)
	return ip, nil
}
