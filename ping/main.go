package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	timeout int64
	size    int
	count   int
	desIP   string
)

type ICMPHeader struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}

func GetCmdArgs() {
	flag.Int64Var(&timeout, "w", 1000, "发送超时时间")
	flag.IntVar(&size, "l", 32, "发送字节数")
	flag.IntVar(&count, "n", 4, "发送次数")
	flag.Parse()
	desIP = flag.Arg(0)
	if desIP == "" {
		panic("必须有目的地址")
	}
}

func GetCheckSum(data []byte) uint16 {
	var sum uint32 = 0
	i := len(data) - 1

	for i > 0 {
		sum += uint32(data[i-1])<<8 + uint32(data[i])
		i -= 2
	}

	if i == 0 {
		sum += uint32(data[0])
	}

	top16 := sum >> 16
	for top16 != 0 {
		// 高16位 + 低 16位
		sum = top16 + uint32(uint16(sum))
		top16 = sum >> 16
	}
	return uint16(^sum)
}

func Bytes2IP(data []byte) string {
	sb := strings.Builder{}
	for i := 0; i < 3; i++ {
		sb.WriteString(strconv.Itoa(int(data[i])))
		sb.WriteByte('.')
	}
	sb.WriteString(strconv.Itoa(int(data[3])))
	return sb.String()
}
func Ping() {
	coon, err := net.DialTimeout("ip:icmp", desIP, time.Duration(10000)*time.Millisecond)
	if err != nil {
		panic(err)
	}

	defer coon.Close()
	data := make([]byte, size, size)
	icmpHeader := &ICMPHeader{
		8,
		0,
		0,
		3,
		0,
	}

	buffer := bytes.Buffer{}
	err = binary.Write(&buffer, binary.BigEndian, icmpHeader)
	if err != nil {
		panic(err.Error())
	}

	buffer.Write(data)
	pckData := buffer.Bytes()
	checkSum := GetCheckSum(pckData)
	pckData[2] = byte(checkSum >> 8)
	pckData[3] = byte(checkSum)
	resBuf := make([]byte, 30+size)
	fmt.Printf("Pinging %s with %d bytes of data :\n", desIP, size)
	for ; count > 0; count-- {
		before := time.Now().UnixMilli()
		_, err = coon.Write(pckData)
		if err != nil {
			panic(err.Error())
		}
		err = coon.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
		if err != nil {
			panic(err.Error())
		}
		n, err := coon.Read(resBuf)
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			fmt.Printf("Request timed out \n")
			continue
		}
		after := time.Now().UnixMilli()
		fmt.Printf("Reply from %s , data bytes = %d , time =%3d (ms) , TTL = %v \n", Bytes2IP(resBuf[12:16]), n-28, after-before, uint8(resBuf[8]))
	}

}

func main() {
	GetCmdArgs()
	Ping()
}
