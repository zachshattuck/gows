# `gows`
Simple WebSocket ([RFC6455](https://datatracker.ietf.org/doc/rfc6455/)) library in Go

**What it offers:**
- Simple way to upgrade an HTTP connection to a WebSocket connection
- Simple way to serialize and deserialize individual WebSocket frames

**What it doesn't do:**
- Doesn't handle message fragmentation, but you can do that yourself by reading `Fin` and `Opcode`. For more information, see section 5.4 of [RFC6455](https://datatracker.ietf.org/doc/rfc6455/)
- Doesn't automatically respond to PING frames.

## Installation
`go get github.com/zachshattuck/gows`

## Example Usage
```go
import (
	"github.com/zachshattuck/gows"
)

func main() {

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Failed to `net.Listen`: ", err)
		os.Exit(1)
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Failed to `Accept` connection: ", err)
		os.Exit(1)
	}

	// Will `Read` from the connection and send a `101 Switching Protocols` response
	// if valid, otherwise sends a `400 Bad Request` response.
	err := gows.UpgradeConnection(&conn, buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to upgrade: ", err)
		os.Exit(1)
	}

	// Listen for WebSocket frames
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to read: ", err)
			break
		}

		frame, err := gows.DeserializeWebSocketFrame(buf[:n])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to deserialize frame: ", err)
			continue
		}

		switch frame.Opcode {
		case gows.WS_OP_TEXT: // Handle text frame..
		case gows.WS_OP_BIN: // Handle binary frame..
		case gows.WS_OP_PING:
			fmt.Println("Ping frame, responding with pong...")
			pongFrame := gows.SerializeWebSocketFrame(gows.WebSocketFrame{
				Fin:  1,
				Rsv1: 0, Rsv2: 0, Rsv3: 0,
				Opcode:   gows.WS_OP_PONG,
				IsMasked: 0,
				MaskKey:  [4]byte{},
				Payload:  frame.Payload,
			})
			conn.Write(pongFrame)
		}

	}

}
```