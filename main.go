package main

// Members
// Nipun Wahi 2018A7PS0966H
// Hrithik Kulkarni 2018A7PS0278H
// Ameetesh Sharma 2018A7PS0167H
// Mir Ameen Mohideen 2018A7PS0487H
// Nielless Acharya 2018A7PS0207H

import (
	"Share/client"
	"Share/server"
	"flag"
	"fmt"
	"net"
	_ "net/http/pprof"
)

func main() {
	var filePath string
	var Port string

	flag.StringVar(&filePath, "PATH", "//", "abc/d/e")
	flag.StringVar(&Port, "PORT", ":4444", "1234")
	flag.Parse()
	if filePath != "//" {
		addr, err := net.ResolveUDPAddr("udp4", Port)
		if err != nil {
			panic(err)
		}
		server.Send(filePath, addr)
	} else {
		addr, err := net.ResolveUDPAddr("udp4", Port)
		if err != nil {
			panic(err)
		}
		fmt.Println("addr of server:", addr)
		client.Recv(addr)
	}

}
