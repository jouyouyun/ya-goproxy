package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

/*
  TCP Header Format:
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |          Source Port          |       Destination Port        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                        Sequence Number                        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Acknowledgment Number                      |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Data |           |U|A|P|R|S|F|                               |
   | Offset| Reserved  |R|C|S|S|Y|I|            Window             |
   |       |           |G|K|H|T|N|N|                               |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           Checksum            |         Urgent Pointer        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Options                    |    Padding    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                             data                              |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

  Currently defined options include (kind indicated in octal):
      Kind     Length    Meaning
      ----     ------    -------
       0         -       End of option list.
       1         -       No-Operation.
       2         4       Maximum Segment Size.

   kind=2: Maximum Segment Size
        +--------+--------+---------+--------+
        |00000010|00000100|   max seg size   |
        +--------+--------+---------+--------+
         Kind=2   Length=4

  Detail see: https://tools.ietf.org/html/rfc793#page-15

  Max playload: 1480-20=1460
*/

type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte // kind=2, 16 bits
}
type TCPOptions []*TCPOption

type TCPHeader struct {
	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4 bits
	Reserved    uint8 // 6 bits
	URG         byte
	ACK         byte
	PSH         byte
	RST         byte
	SYN         byte
	FIN         byte
	Window      uint16
	Checksum    uint16
	Urgent      uint16
	Options     TCPOptions // variable
	Padding     []byte     // variable, ensure that the TCP header ends and data begins on a 32 bit boundary
}

func UnmarshalTCPHeader(data []byte) *TCPHeader {
	var header TCPHeader
	var r = bytes.NewReader(data)
	binary.Read(r, binary.BigEndian, &header.Source)
	binary.Read(r, binary.BigEndian, &header.Destination)
	binary.Read(r, binary.BigEndian, &header.SeqNum)
	binary.Read(r, binary.BigEndian, &header.AckNum)
	var tmp uint16
	binary.Read(r, binary.BigEndian, &tmp)
	header.DataOffset = uint8(tmp >> 12)     // top 4 bits
	header.Reserved = uint8(tmp >> 6 & 0x3f) // middle 6 bits
	header.URG = byte(tmp & 0x20)
	header.ACK = byte(tmp & 0x10)
	header.PSH = byte(tmp & 0x8)
	header.RST = byte(tmp & 0x4)
	header.SYN = byte(tmp & 0x2)
	header.FIN = byte(tmp & 0x1)
	binary.Read(r, binary.BigEndian, &header.Window)
	binary.Read(r, binary.BigEndian, &header.Checksum)
	binary.Read(r, binary.BigEndian, &header.Urgent)
	// TODO: handle options
	return &header
}

func (header *TCPHeader) String() string {
	var flags []string
	if header.URG != byte(0) {
		flags = append(flags, "U")
	}
	if header.ACK != byte(0) {
		flags = append(flags, "A")
	}
	if header.PSH != byte(0) {
		flags = append(flags, "P")
	}
	if header.RST != byte(0) {
		flags = append(flags, "R")
	}
	if header.SYN != byte(0) {
		flags = append(flags, "S")
	}
	if header.FIN != byte(0) {
		flags = append(flags, "F")
	}
	var flag = strings.Join(flags, "|")
	// TODO: handle options
	return fmt.Sprintf("Source=%v Destination=%v SeqNum=%v AckNum=%v DataOffset=%v Reserved=%d Flags=[%s] Window=%v Checksum=%v Urgent=%v\n",
		header.Source, header.Destination, header.SeqNum, header.AckNum,
		header.DataOffset, header.Reserved, flag, header.Window,
		header.Checksum, header.Urgent)
}

func (header *TCPHeader) Marshal() []byte {
	var buf = new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, header.Source)
	binary.Write(buf, binary.BigEndian, header.Destination)
	binary.Write(buf, binary.BigEndian, header.SeqNum)
	binary.Write(buf, binary.BigEndian, header.AckNum)
	var ctrl uint8
	ctrl = uint8(header.URG)<<5 | uint8(header.ACK)<<4 |
		uint8(header.PSH)<<3 | uint8(header.RST)<<2 |
		uint8(header.SYN)<<1 | uint8(header.FIN)
	var tmp uint16
	tmp = uint16(header.DataOffset)<<12 | // top 4 bites
		uint16(header.Reserved)<<6 | // middle 6 bits
		uint16(ctrl) // bottom 6 bits
	binary.Write(buf, binary.BigEndian, tmp)
	binary.Write(buf, binary.BigEndian, header.Window)
	binary.Write(buf, binary.BigEndian, header.Checksum)
	binary.Write(buf, binary.BigEndian, header.Urgent)
	for _, option := range header.Options {
		binary.Write(buf, binary.BigEndian, option.Kind)
		if option.Length < 1 {
			continue
		}
		binary.Write(buf, binary.BigEndian, option.Length)
		binary.Write(buf, binary.BigEndian, option.Data)
	}
	out := buf.Bytes()
	// padding to min tcp header size, which is 20 bytes (5 32-bit words)
	padding := 20 - len(out)
	for i := 0; i < padding; i++ {
		out = append(out, 0)
	}
	return out
}
