package main

import (
	"log"
	"net"
)

func main() {
	local := getOutboundIP()
	conn, err := net.Dial("tcp", local.String()+":8379")
	if err != nil {
		StartServer()
		conn, err = net.Dial("tcp", ":8379")
	}

}

func StartServer(ip string) {
	server, err := net.Listen("tcp", ip)
	if err != nil {
		log.Fatalln(err)
	}

}

func proxy(from)

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
