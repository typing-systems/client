package main

import (
	"log"
	"net"
	"os"
	"time"
)

const (
	HOST = "localhost"
	PORT = "9001"
	TYPE = "tcp"
)

func main() {
	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		go handleIncomingRequest(conn)
	}
}

func handleIncomingRequest(conn net.Conn) {
	// storing incoming data
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// respond
	time := time.Now().Format("15:04:05 02/Jan/2006")
	conn.Write([]byte("Message received at "))
	conn.Write([]byte(time + "\n"))
}
