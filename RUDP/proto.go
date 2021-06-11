package RUDP

import (
	val "Share/VARIABLES"
	"Share/dictionary"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	_ "net/http/pprof"
	"sort"
	"time"
)

const (
	synFlag uint8 = iota
	ackFlag
	synAck
	finFlag
	sndFlag
)

var pktSIZE int = val.PKTSIZE

type Packet struct {
	seq    uint32
	ackseq uint32
	flag   uint8
	size   uint32
	data   []byte
}

type RUDP struct {
	addr             *net.UDPAddr
	conn             *net.UDPConn
	seq              uint32
	ackseq           uint32
	recvpkt          *Packet
	lastReadseq      uint32
	recvChn          chan *Packet
	ackRcv           chan bool
	isSync           bool
	state            int
	isListen         bool
	lastack          uint32
	synackdone       bool
	lastsendDone     bool
	buffer           map[uint32][]byte
	threadSafeBuffer dictionary.Dictionary
	syndone          bool
}

type Response struct {
	pkt *Packet
	err error
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NewRUDP(addr *net.UDPAddr, conn *net.UDPConn, isServer bool) *RUDP {
	return &RUDP{
		addr:             addr,
		conn:             conn,
		seq:              0,
		ackseq:           0,
		recvpkt:          nil,
		lastReadseq:      0,
		recvChn:          make(chan *Packet),
		ackRcv:           make(chan bool, 0),
		isSync:           false,
		state:            0,
		isListen:         isServer,
		lastack:          0,
		synackdone:       false,
		lastsendDone:     true,
		buffer:           make(map[uint32][]byte),
		threadSafeBuffer: dictionary.Dictionary{},
		syndone:          false,
	}
}

func (c *RUDP) send(data []byte) error {
	if c.isListen {
		// fmt.Println("Server sending to", c.addr, getPktType(getPkt(data)))
		_, err := c.conn.WriteToUDP(data, c.addr)
		if err != nil {
			return err
		}
	} else {
		// fmt.Println("Sending", getPkt(data))
		// fmt.Println("Client sending to", c.addr, getPktType(getPkt(data)), getPkt(data).seq)
		_, err := c.conn.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RUDP) recv() *Packet {
	buf := make([]byte, 2*pktSIZE)
	if true {
		n, addr, err := c.conn.ReadFromUDP(buf)
		c.addr = addr
		// fmt.Println("In recv func", c.addr)
		if err != nil {
			panic(err)
		}
		return getPkt(buf[:n])
	} else {
		n, err := c.conn.Read(buf)
		if err != nil {
			panic(err)
		}
		return getPkt(buf[:n])
	}
}

func (c *RUDP) Reciever() {
	// ss := time.Now()
	// en := time.Since(ss)
	for {
		// fmt.Println("LISTENING")
		pkt := c.recv()
		// fmt.Println(getPktType(pkt))
		if !verifyPkt(pkt) {
			fmt.Println("here ")
			continue
		}

		if pkt.flag == synFlag && !c.synackdone && !c.syndone {
			c.lastReadseq = pkt.seq
			c.ackseq = 0
			c.syndone = true
			go c.sendsynackrept()
		}

		if pkt.flag == synAck {
			if !c.synackdone {
				c.lastReadseq = pkt.seq
				c.ackseq = 0
				c.send(makeAck(0, pkt.seq+1))
				c.synackdone = true
				c.ackRcv <- true
			} else {
				c.send(makeAck(0, pkt.seq+1))
			}

		}

		if pkt.flag == ackFlag {
			// fmt.Println("ACK RECIEVED FOR ACK ", pkt.ackseq-1)
			if pkt.ackseq > c.ackseq {
				c.ackseq = pkt.ackseq
				arr := make([]uint32, 0)
				for _, k := range c.threadSafeBuffer.GetKeys() {
					if k < pkt.ackseq {
						arr = append(arr, k)
					}
				}
				for _, v := range arr {
					c.threadSafeBuffer.Remove(v)
				}
				if c.threadSafeBuffer.Size() == 0 {
					// en = time.Since(ss)
					c.ackRcv <- true
					// ss = time.Now()
					// fmt.Println("Time for final ack", en)
				}

			} else {
				if c.threadSafeBuffer.Exist(pkt.ackseq) {
					c.send(c.threadSafeBuffer.Get(pkt.ackseq))
				}
			}
		}

		if pkt.flag == sndFlag {
			// fmt.Println(pkt.seq, c.lastReadseq, "IN CLIENT SEND")
			if pkt.seq == c.lastReadseq+1 {
				c.lastReadseq = pkt.seq
				c.send(makeAck(c.lastReadseq+1, pkt.seq+1))
				c.recvChn <- pkt
			} else {
				c.send(makeAck(c.lastReadseq+1, c.lastReadseq+1))
			}
		}

	}
}

func (c *RUDP) ResendReliable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	arr := make([]int, 0)
	for _, v := range c.threadSafeBuffer.GetKeys() {
		// fmt.Println("sequence in buffer", k)
		arr = append(arr, int(v))
	}
	sort.Ints(arr)
	// ss := time.Now()
	for x := range arr {
		c.send(c.threadSafeBuffer.Get(uint32(arr[x])))
	}
	// en := time.Since(ss)
	// fmt.Println("send took ", en)
	select {
	case <-c.ackRcv:
		return true
	case <-ctx.Done():
		// fmt.Println("ACKS NOT RCVED ")
		return false
	}
}

func (c *RUDP) sendReliable(data []byte) {
	size := len(data)
	chunks := size / pktSIZE
	if size%pktSIZE != 0 {
		chunks++
	}
	// ss := time.Now()
	for i := 0; i < chunks; i++ {
		c.seq++
		dsize := min(pktSIZE, len(data)-pktSIZE*i)
		buf := makePkt(c.seq, uint32(dsize), sndFlag, data[i*pktSIZE:min((i+1)*pktSIZE, len(data))], 0)
		c.threadSafeBuffer.Add(c.seq, buf)
	}
	// fmt.Println("final seq num", c.seq)
	// en := time.Since(ss)
	// fmt.Println("making packets took ", en)
	var done bool = false
	var calling int = 0
	for {
		if done {
			break
		}
		calling++
		// ss := time.Now()
		done = c.ResendReliable()
		// en := time.Since(ss)
		// fmt.Println("time in resend ", en)
		// fmt.Println("DONE ", done)
		// fmt.Println("called ", calling)
	}
}

func (c *RUDP) Read(buf []byte) (int, error) {
	// ss := time.Now()
	pkt := <-c.recvChn
	// en := time.Since(ss)
	// fmt.Println("channel took", en)
	if len(pkt.data) > len(buf) {
		// panic("Buffer size small in read")
		return 0, fmt.Errorf("BUFFER SIZE TOO SMALL")
	}
	for i := 0; i < len(pkt.data); i++ {
		buf[i] = pkt.data[i]
	}

	return len(pkt.data), nil
}

func (c *RUDP) Write(data []byte) {
	for !c.synackdone {

	}
	c.sendReliable(data)
	// fmt.Println("SENT RELIABLE DONE")
}

func (c *RUDP) Connect(isDone chan bool) {
	go c.Reciever()
	c.handshake()
	isDone <- true
	for {

	}

}
func (c *RUDP) sendsynackrept() {
	for {
		Resp := c.sendSynAck()
		if Resp.err == nil {
			c.synackdone = true
			break
		}
	}
}

func (c *RUDP) sendSynAck() *Response {
	buf := makeSynAck(0, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go c.send(buf)
	select {
	case <-ctx.Done():
		fmt.Println("TIMEOUT IN SENDSYNACK")
		return &Response{nil, fmt.Errorf("context timeout, ran out of time in synack")}
	case <-c.ackRcv:
		return &Response{c.recvpkt, nil}
	}
}

func (c *RUDP) sendSyn() *Response {
	buf := makeSyn()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go c.send(buf)
	select {
	case <-ctx.Done():
		fmt.Println("TIMEOUT IN SENDSYN")
		return &Response{nil, fmt.Errorf("context timeout, ran out of time")}
	case <-c.ackRcv:
		return &Response{c.recvpkt, nil}
	}
}

func (c *RUDP) handshake() bool {
	fmt.Println("Started Handshake")
	resp := c.sendSyn()
	if resp.err != nil {
		fmt.Println(resp.err.Error())
		return c.handshake()
	} else {
		// c.send(makeAck(0, resp.pkt.seq+1))
		return true
	}
}

func makeSyn() []byte {
	return makePkt(0, 0, synFlag, []byte(""), 0)
}

func makeAck(seq, ackseq uint32) []byte {
	return makePkt(seq, 0, ackFlag, []byte(""), ackseq)
}

func makeSynAck(seq, ackseq uint32) []byte {
	return makePkt(seq, 0, synAck, []byte(""), ackseq)
}

func makesendPkt(seq, ackseq uint32, data []byte) []byte {
	return makePkt(seq, uint32(len(data)), sndFlag, data, 0)
}

func makePkt(seq uint32, size uint32, flag uint8, data []byte, ackseq uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, flag)
	binary.Write(buf, binary.BigEndian, seq)
	binary.Write(buf, binary.BigEndian, ackseq)
	binary.Write(buf, binary.BigEndian, size)
	buf.Write(data)
	return buf.Bytes()
}

func getPkt(buf []byte) *Packet {
	pkt := new(Packet)
	buffer := new(bytes.Buffer)
	buffer.Write(buf)
	binary.Read(buffer, binary.BigEndian, &pkt.flag)
	binary.Read(buffer, binary.BigEndian, &pkt.seq)
	binary.Read(buffer, binary.BigEndian, &pkt.ackseq)
	binary.Read(buffer, binary.BigEndian, &pkt.size)
	pkt.data = make([]byte, pkt.size)
	buffer.Read(pkt.data)
	return pkt
}

func getPktType(pkt *Packet) string {
	if pkt.flag == synAck {
		return "SYNACK"
	}
	if pkt.flag == synFlag {
		return "SYN"
	}
	if pkt.flag == sndFlag {
		return "SEND"
	}
	if pkt.flag == ackFlag {
		return "ACK"
	}
	return "OTH"
}

func verifyPkt(pkt *Packet) bool {
	if uint32(len(pkt.data)) != pkt.size {
		return false
	}
	return true
}
