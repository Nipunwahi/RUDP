package client

import (
	rudp "Share/RUDP"
	val "Share/VARIABLES"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type FileInfo struct {
	name string
	size int64
}

var completed chan bool

func makeHeader(name string, datalen int64) []byte {
	buf := new(bytes.Buffer)
	var length int64 = int64(len(name))
	binary.Write(buf, binary.BigEndian, datalen)
	binary.Write(buf, binary.BigEndian, length)
	buf.Write([]byte(name))
	return buf.Bytes()
}

func decodeHeader(buf []byte) *FileInfo {
	buffer := new(bytes.Buffer)
	buffer.Write(buf)
	var datalen int64
	var namelen int64
	binary.Read(buffer, binary.BigEndian, &datalen)
	binary.Read(buffer, binary.BigEndian, &namelen)
	namebuf := make([]byte, namelen)
	buffer.Read(namebuf)
	name := string(namebuf)
	return &FileInfo{
		name: name,
		size: datalen,
	}
}

func read(RUDP *rudp.RUDP) {
	buf := make([]byte, 2*val.PKTSIZE)
	flag := false
	var file *os.File
	var length int64
	sum := 0
	for {
		n, err := RUDP.Read(buf)
		if n == 0 {
			return
		}
		if err != nil {
			return
		}
		if flag {
			file.Write(buf[:n])
			sum += n
			if int(length) == sum {
				break
			}
		}
		if !flag {
			info := decodeHeader(buf[:n])
			names := "shared" + info.name
			length = info.size
			names = strings.ReplaceAll(names, "\x00", "")
			filecpy, err := os.OpenFile(names, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			file = filecpy
			flag = true
		}

		// fmt.Println(string(buf[:n]))
	}
	completed <- true
}

func Recv(addr *net.UDPAddr) {
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		panic(err)
	}
	completed = make(chan bool)
	RUDP := rudp.NewRUDP(addr, conn, false)
	isDone := make(chan bool)
	go RUDP.Connect(isDone)
	go read(RUDP)
	<-isDone
	fmt.Println("COMPLETED HANDSHAKE")
	start := time.Now()
	<-completed
	elapsed := time.Since(start)
	log.Printf("SEND took %s", elapsed)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	<-ctx.Done()
}
