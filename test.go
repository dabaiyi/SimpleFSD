package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:6809")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on :6809")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		log.Printf("Accepted connection from: %s", conn.RemoteAddr())
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	log.Printf("Received %d bytes: %s", n, string(buf[:n]))
}
