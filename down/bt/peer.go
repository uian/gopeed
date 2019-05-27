package bt

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

type MessageType byte

const (
	ProtocolIdentifier       = "BitTorrent protocol"
	ProtocolIdentifierLength = 0x13

	Choke MessageType = iota
	Unchoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
)

type Peer struct {
	IP   uint32
	Port uint16
}

func (peer *Peer) Address() string {
	bts := make([]byte, 4)
	binary.BigEndian.PutUint32(bts, peer.IP)
	return net.IP(bts).String() + ":" + strconv.Itoa(int(peer.Port))
}

type Handshake struct {
	Pstrlen  byte
	Pstr     [19]byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

type Message struct {
	Length  uint32
	Type    byte
	Payload []byte
}

func NewHandshake(reserved [8]byte, infoHash [20]byte, peerID [20]byte) *Handshake {
	var arr [ProtocolIdentifierLength]byte
	copy(arr[:], ProtocolIdentifier)
	return &Handshake{
		Pstrlen:  ProtocolIdentifierLength,
		Pstr:     arr,
		Reserved: reserved,
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (handshake *Handshake) encode() ([]byte, error) {
	writer := new(bytes.Buffer)
	err := binary.Write(writer, binary.BigEndian, handshake)
	if err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func (handshake *Handshake) decode(buf []byte) error {
	reader := bytes.NewReader(buf)
	err := binary.Read(reader, binary.BigEndian, handshake)
	if err != nil {
		return err
	}
	return nil
}

func (peer *Peer) DoDownload(metaInfo *MetaInfo, peerId [20]byte) error {
	if peer.Port == 0 {
		return fmt.Errorf("error port %d", peer.Port)
	}
	conn, err := net.Dial("tcp", peer.Address())
	if err != nil {
		return err
	}
	defer conn.Close()
	err = doHandshake(peer, metaInfo, peerId, conn)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	scanner.Split(splitMessage)
	for scanner.Scan() {
		buf := scanner.Bytes()
		message := decodeMessage(buf)
		// keep-alive message
		if message.Length == 0 {

		} else {
			switch MessageType(buf[4]) {
			case Choke:
				break
			case Unchoke:
				break
			case Interested:
				break
			case NotInterested:
				break
			case Have:
				break
			case Bitfield:
				break
			case Request:
				break
			case Piece:
				break
			case Cancel:
				break
			}
		}
	}

	/*var (
		// This client is choking the peer
		amChocking = true
		// This client is interested in the peer
		amInterested = false
		// Peer is choking this client
		peerChocking = true
		// Peer is interested in this client
		peerInterested = false
	)

	// Receive message
	var cacheBuf []byte
	buf := make([]byte, 8192)
	for {
		scanner := bufio.NewScanner(conn)
		n, err := conn.Read(buf)
		if n > 0 {
			if len(cacheBuf) > 0 {
				buf = append(cacheBuf, buf...)
			}
			msg, err := decodeMessage(buf, n)
			if err != nil {
				if err == NMD {
					cacheBuf = make([]byte, n)
					copy(cacheBuf, buf[0:n])
				}
				return err
			}

			msgLength := binary.BigEndian.Uint32(buf)
			// keep-alive message
			if msgLength == 0 {

			} else {
				switch MessageType(buf[4]) {
				case Choke:
					peerChocking = true
					break
				case Unchoke:
					peerChocking = false
					break
				case Interested:
					peerInterested = true
					break
				case NotInterested:
					peerInterested = false
					break
				case Have:
					break
				case Bitfield:
					break
				case Request:
					break
				case Piece:
					break
				case Cancel:
					break

				}

			}

		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}*/
	return nil
}

// Handshake of Peer wire protocol
// Per https://wiki.theory.org/index.php/BitTorrentSpecification#Handshake
func doHandshake(peer *Peer, metaInfo *MetaInfo, peerId [20]byte, conn net.Conn) error {
	reserved := [8]byte{}
	reserved[5] = 0x10
	reserved[6] = 0x0
	reserved[7] = 0x5
	handshakeReq := NewHandshake(reserved, metaInfo.InfoHash, peerId)
	buf, err := handshakeReq.encode()
	if err != nil {
		return err
	}
	conn.Write(buf)

	var read [68]byte
	_, err = io.ReadFull(conn, read[:])
	if err != nil {
		return err
	}
	handshakeRes := new(Handshake)
	err = handshakeRes.decode(read[:])
	if err != nil {
		return err
	}
	if handshakeRes.InfoHash != handshakeReq.InfoHash {
		return err
	}
	return nil
}

func splitMessage(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF && len(data) > 4 {
		length := int(binary.BigEndian.Uint32(data))
		if len(data)-4 >= length {
			return length + 4, data[:length+4], nil
		}
	}
	return
}

func decodeMessage(buf []byte) *Message {
	msg := &Message{}
	msg.Length = binary.BigEndian.Uint32(buf)
	// keep-alive message
	if msg.Length > 0 {
		msg.Type = buf[4]
		msg.Payload = buf[5:]
	}
	return msg
}

func doBitfield(peer *Peer, metaInfo *MetaInfo, conn net.Conn) {

}
