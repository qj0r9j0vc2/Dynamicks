package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	port string
)

const IP_URL = "https://api.ipify.org?format=text"

func init() {
	flag.StringVar(&port, "port", "8992", "Port To Use")
}

func main() {

	local := getPubIP()

	fmt.Println(local)

	addr := fmt.Sprintf(local + ":" + port)

	fmt.Println("Connect To : ", addr)
	conn, err := net.Dial("tcp", addr)
	wg := sync.WaitGroup{}

	if err != nil {
		fmt.Println("There's no Hosting Server!")
		wg.Add(1)
		go HostServer(&wg)
		wg.Wait()
		conn, err = net.Dial("tcp", ":8992")
		if err != nil {
			fmt.Println("Cannot Connect To Server!!!")
		}
	}

	wg.Add(1)
	go RunClient(conn)
	wg.Wait()
}

func RunClient(conn net.Conn) {

	go func(c net.Conn) {
		recv := make([]byte, 4096)

		for {
			n, err := c.Read(recv)
			if err != nil {
				fmt.Println("Failed to Read data : ", err)
				break
			}
			fmt.Println("[Received]: ", string(recv[:n]))
		}
	}(conn)

	//메시지 전송 함수 작성해야함.
	for {
		fmt.Printf("[SEND]: ")
		var msg string
		fmt.Scan(&msg)
		_, err := conn.Write([]byte(msg))
		time.Sleep(1 * time.Millisecond)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}

func HostServer(wg *sync.WaitGroup) {
	listen, err := net.Listen("tcp", ":"+port)
	fmt.Println("Starting Server on:", port)

	if err != nil {
		log.Fatalln(err)
	}
	defer listen.Close()

	wg.Done()

	for {
		conn, err := listen.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Cannot connect with Client: ", err.Error())
			continue
		}
		for {
			err = proxy(conn)
			if err != nil {
				break
			}
		}
	}
}

func proxy(conn net.Conn) error {
	received := make([]byte, 4096)
	n, err := conn.Read(received)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Network Connection Failed...")
		} else {
			fmt.Println("Error Occurred while Receiving Data")
			return err
		}
	}
	if n > 0 {
		_, err := conn.Write(received[:n])
		if err != nil {
			fmt.Println("Cannot Write Data...! ", err.Error())
			return err
		}
	}

	return nil
}

func getPubIP() string {
	fmt.Printf("Getting IP address from  ipify ...\n")
	resp, err := http.Get(IP_URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	ip = []byte(string(ip))
	fmt.Printf("My IP is: %s\n", ip)
	return string(ip)
}
