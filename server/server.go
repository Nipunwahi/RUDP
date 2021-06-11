package server

import (
	rudp "Share/RUDP"
	val "Share/VARIABLES"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type FileInfo struct {
	name string
	size int64
}

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
	namebuf := make([]byte, datalen)
	buffer.Read(namebuf)
	name := string(namebuf)
	return &FileInfo{
		name: name,
		size: datalen,
	}
}

func Send(filename string, addr *net.UDPAddr) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}
	RUDP := rudp.NewRUDP(addr, conn, true)
	go RUDP.Reciever()
	fmt.Println("Listening on :", addr)
	stat, err := file.Stat()
	length := stat.Size()
	name := stat.Name()
	header := makeHeader(name, length)
	RUDP.Write(header)

	buf := make([]byte, val.PKTSIZE*val.WINDOWSIZE)
	sum := 0
	prev := 0
	start := time.Now()
	for {
		n, err := file.Read(buf)
		sum += n
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		RUDP.Write(buf[:n])
		if (sum*100)/int(length) > prev {
			fmt.Println("Percent Done", prev)
			prev = (sum * 100) / int(length)
		}
	}
	elapsed := time.Since(start)
	log.Printf("SEND took %s", elapsed)
}
