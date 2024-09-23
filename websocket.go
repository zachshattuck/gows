package gows

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
)

// https://datatracker.ietf.org/doc/rfc6455

/* -> request
REQUIRED:
- verb must be GET
- HTTP version must be at least 1.1
- "Host"
- "Upgrade: websocket"
- "Connection: Upgrade"
- "Sec-WebSocket-Key"
- "Sec-WebSocket-Version: 13" (must be 13?)
- "Origin" if from browser
-
OPTIONAL:
- "Sec-WebSocket-Protocol"
- "Sec-WebSocket-Extensions"
-
*/

/* <- response
REQUIRED:
- "HTTP{version} 101 Switching Protocols"
- "Upgrade: websocket"
- "Connection: Upgrade"
- "Sec-WebSocket-Accept: {base64(sha1(key + magicString))}"

OPTIONAL:
- "Sec-WebSocket-Protocol"
- "Sec-WebSocket-Extensions"
*/

/* <-> framing
client must always mask messages sent, if not server sends back opcode 1002
server never masks

*/

const WS_OP_CONT = 0x0
const WS_OP_TEXT = 0x1
const WS_OP_BIN = 0x2

// 3-7: reserved for future non-control frames

const WS_OP_CLOSE = 0x8
const WS_OP_PING = 0x9
const WS_OP_PONG = 0xA

// B-F reserved for future control frames

const webSocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

/*
Reads the next message and tries to interpret it as a WebSocket upgrade request.
If invalid, will send a 400 Bad Request response.
If valid, will send a 101 Switching Protocols response.
*/
func UpgradeConnection(conn *net.Conn, buf []byte) error {
	n, err := (*conn).Read(buf)
	if err != nil {
		return errors.New("failed to read from connection")
	}

	isValid, err := isValidUpgradeRequest(buf[:n])
	if err != nil || !isValid {
		sendBadRequestResponse(conn)
		return err
	}

	httpWebSocketKey, err := getHttpParam(buf[:n], "Sec-WebSocket-Key")
	if err != nil {
		sendBadRequestResponse(conn)
		return err
	}

	// https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers#server_handshake_response
	sha1Checksum := sha1.Sum([]byte(string(httpWebSocketKey) + webSocketMagicString))
	httpWebSocketAccept := base64.StdEncoding.EncodeToString(sha1Checksum[:])

	// TODO: This should respond with the same HTTP and Sec-WebSocket-Version that the client set, instead of a hardcoded version
	_, err = (*conn).Write([]byte("HTTP/1.1 101 Switching Protocols\r\nSec-WebSocket-Version: 13\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + httpWebSocketAccept + "\r\n\r\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to write upgrade response:", err)
		return err
	}

	return nil
}

type WebSocketFrame struct {
	Fin      byte
	Rsv1     byte
	Rsv2     byte
	Rsv3     byte
	Opcode   uint8
	IsMasked byte
	MaskKey  [4]byte
	Payload  []byte
}

func DeserializeWebSocketFrame(buf []byte) (WebSocketFrame, error) {

	// cannot be a valid websocket frame without 2 bytes
	if len(buf) < 2 {
		return WebSocketFrame{}, errors.New("frame less than 2 bytes")
	}
	// fmt.Println("length of WebSocket frame:", len(buf))
	// byte 1: FIN, RSV1-3, OPCODE
	// byte 2: mask, 7-bit length

	byte1 := buf[0]
	fin := byte1 >> 7
	rsv1 := (byte1 & 0x40) >> 6
	rsv2 := (byte1 & 0x20) >> 5
	rsv3 := (byte1 & 0x10) >> 4
	opcode := byte1 & 0x0F

	byte2 := buf[1]
	isMasked := byte2 >> 7
	payloadLen := uint64(byte2 & 0x7F)

	maskKeyStartIdx := 2

	// Calculate extended payload length...
	if payloadLen == 126 {
		if len(buf) < 4 {
			return WebSocketFrame{}, errors.New("frame not long enough to read extended payload length")
		}
		s := buf[2:4]
		payloadLen = (uint64(s[0]) << 8) | uint64(s[1])
		maskKeyStartIdx = 4
	} else if payloadLen == 127 {
		if len(buf) < 10 {
			return WebSocketFrame{}, errors.New("frame not long enough to read extended payload length")
		}
		s := buf[2:10]
		payloadLen = (uint64(s[0]) << 56) |
			(uint64(s[1]) << 48) |
			(uint64(s[2]) << 40) |
			(uint64(s[4]) << 32) |
			(uint64(s[5]) << 24) |
			(uint64(s[6]) << 16) |
			(uint64(s[7]) << 8) |
			(uint64(s[8]) << 0)
		maskKeyStartIdx = 10
	}

	payloadStartIdx := 0
	var maskKey [4]byte
	if isMasked == 1 {
		if len(buf) < maskKeyStartIdx+4 {
			return WebSocketFrame{}, errors.New("frame not long enough to read mask")
		}
		// Is it necessary to copy the maskKey out of the original buffer to put into the deserialized struct?
		// idk
		copy(maskKey[:], buf[maskKeyStartIdx:maskKeyStartIdx+4])
		payloadStartIdx = maskKeyStartIdx + 4
	}

	if len(buf) < payloadStartIdx+int(payloadLen) {
		return WebSocketFrame{}, errors.New("frame not long enough for payload of given length")
	}

	payload := buf[payloadStartIdx : payloadStartIdx+int(payloadLen)]

	if isMasked == 1 {
		for i, bite := range payload {
			payload[i] = bite ^ maskKey[i%4]
		}
	}

	return WebSocketFrame{
		Payload:  payload,
		Fin:      fin,
		Rsv1:     rsv1,
		Rsv2:     rsv2,
		Rsv3:     rsv3,
		IsMasked: isMasked,
		MaskKey:  maskKey,
		Opcode:   opcode}, nil
}

func SerializeWebSocketFrame(data WebSocketFrame) []byte {

	var frame []byte

	/** Size of the frame, not including the payload. */
	frameHeaderLen := uint64(2)
	if data.IsMasked == 1 {
		frameHeaderLen += 4
	}

	byte1 := (data.Fin << 7) |
		((data.Rsv1 << 6) & 0x40) |
		((data.Rsv2 << 5) & 0x20) |
		((data.Rsv3 << 4) & 0x10) |
		data.Opcode&0x0F

	byte2 := ((data.IsMasked & 0x01) << 7)

	// Note: This will truncate any value larger than a 64-bit unsigned integer.
	payloadLen := uint64(len(data.Payload))

	// Create the frame based on how many bytes we need for payload length
	if payloadLen >= 126 {
		if payloadLen < math.MaxUint16 { //16-bit extended payload length
			byte2 |= (0x7E)
			frameHeaderLen += 2
			frame = make([]byte, frameHeaderLen+payloadLen)
			binary.BigEndian.PutUint16(frame[2:], uint16(payloadLen))
		} else { // 64-bit extended payload length
			byte2 |= (0x7F)
			frameHeaderLen += 8
			frame = make([]byte, frameHeaderLen+payloadLen)
			binary.BigEndian.PutUint64(frame[2:], payloadLen)
		}
	} else {
		byte2 |= (byte(payloadLen) & 0x7F)
		frame = make([]byte, frameHeaderLen+payloadLen)
	}

	frame[0] = byte1
	frame[1] = byte2

	// Copy the payload into the frame
	copy(frame[frameHeaderLen:], data.Payload)

	// Copy the mask key into the frame and XOR-encrypt the payload
	if data.IsMasked == 1 {
		copy(frame[frameHeaderLen-4:], data.MaskKey[:])
		for i := range frame[frameHeaderLen:] {
			frame[frameHeaderLen:][i] ^= data.MaskKey[i%4]
		}
	}

	return frame
}
