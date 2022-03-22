package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
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

	local := getOutboundIP()
	addr := fmt.Sprintf(local + ":" + port)

	fmt.Printf("Connect To : %s\n\n", addr)
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	wg := sync.WaitGroup{}

	if err != nil {
		fmt.Println("There's no Hosting Server!\n\n\n")
		wg.Add(1)
		go HostServer(&wg)
		wg.Wait()
		conn, err = net.Dial("tcp", ":8992")
		if err != nil {
			fmt.Println("Cannot Connect To Server!!!")
		}
	}

	wg.Add(1)
	go RunClient(conn, &wg)
	wg.Wait()
}

func RunClient(conn net.Conn, wg *sync.WaitGroup) {

	clientPrintln("Local address: ", conn.LocalAddr())

	var name string
	fmt.Println("Enter your Name: ")
	fmt.Scanln(&name)

	go func(c net.Conn) {
		recv := make([]byte, 4096)

		for {
			n, err := c.Read(recv)
			if err != nil {
				if err == io.EOF {
					serverPrintln("Hosting Server DOWN...!")
					return
				}
				serverPrintln("Failed to Read data : ", err.Error())
				break
			}
			resp := string(recv[:n])
			log.Println(strings.Replace(resp, "["+name+"]", "[ME]", 1))
		}
	}(conn)

	in := bufio.NewReader(os.Stdin)
	for {
		msg := newLineScanln(name, in)
		_, err := conn.Write([]byte(msg))
		time.Sleep(1 * time.Millisecond)

		if err != nil {
			connectPrintln(err.Error())
			wg.Done()
			return
		}

	}

}

func newLineScanln(name string, in *bufio.Reader) string {
	line, err := in.ReadString('\n')
	line = strings.Trim(line, "\n")

	if err != nil {
		fmt.Println("[ERROR] invalid input data")
	}
	return "[" + name + "]: " + line
}

func HostServer(wg *sync.WaitGroup) {
	listen, err := net.Listen("tcp", ":"+port)
	fmt.Printf("Starting Server on: %s\n\n", port)

	if err != nil {
		log.Fatalln(err)
	}
	defer listen.Close()

	wg.Done()
	connList := make([]net.Conn, 0, 100)

	var connCount = 0
	for {
		conn, err := listen.Accept()
		connList = append(connList, conn)

		defer conn.Close()
		if err != nil {
			connectPrintln("Cannot connect with Client: ", err.Error())
			continue
		}
		serverPrintln("New Client Connected!  CLIENT counts: ", connCount+1, "ClientIP: ", conn.RemoteAddr())

		go func() {
			var cnnIdx = connCount
			for {
				err = proxy(conn, &connList)
				if err != nil {
					break
				}
			}

			rmIndexSlice(connList, cnnIdx, connCount)
			connCount--
			serverPrintln("Disconnected...! CLIENT counts: ", connCount)
		}()
		connCount++
	}
}

func rmIndexSlice(slice []net.Conn, idx int, cnnCnt int) []net.Conn {
	if idx == cnnCnt {
		return slice[:idx]
	}
	return append(slice[:idx], slice[idx+1:]...)
}

func proxy(from net.Conn, toList *[]net.Conn) error {
	received := make([]byte, 4096)
	n, err := from.Read(received)
	if err != nil {
		if err == io.EOF {
			serverPrintln("Network Connection Failed... ClientIP: ", from.RemoteAddr())
			return err
		} else {
			connectPrintln("Error Occurred while Receiving Data")
			return err
		}
	}

	if n > 0 {
		for _, to := range *toList {
			//connectLogging("pipeline: ", to.LocalAddr(), " -> ", to.RemoteAddr())
			_, err := to.Write(received[:n])

			if err != nil {
				fmt.Println("Cannot Write Data...! ", err.Error())
				return err
			}
		}
	}

	return nil
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
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

func serverPrintln(a ...interface{}) {
	log.Println("[SERVER] ", a)
}

func connectPrintln(a ...interface{}) {
	log.Println("[CONNECT] ", a)
}

func receivedPrintln(a ...interface{}) {
	log.Println("[RECEIVED]: ", a)
}

func clientPrintln(a ...interface{}) {
	log.Println("[CLIENT] ", a)
}
