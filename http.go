package gows

import (
	"errors"
	"net"
)

/*
Given a param name and a buffer expected to be a valid HTTP request, this function
will return a slice containing the value of that HTTP param, if it is found.
*/
func getHttpParam(buf []byte, paramName string) ([]byte, error) {

	// Read until we match `paramName` completely, NOT including the ":"
	var correctByteCount int = 0
	var valueStartIdx int
	for i, b := range buf {
		if b != paramName[correctByteCount] {
			correctByteCount = 0
			continue
		}

		// Previous character has to be start of buffer or '\n' (as part of CRLF)
		// NOTE: If the user provided a slice that was partway through a request, this could
		// produce wrong results. For example, if there were two params, "Test-Param1: {value}"
		// and "Param1: {value}", and the slice started at the 'P' in "Test-Param1", it could
		// extract that value as if it was just "Param1".
		if correctByteCount == 0 && !(i == 0 || buf[i-1] == '\n') {
			correctByteCount = 0
			continue
		}

		correctByteCount++

		if correctByteCount < len(paramName) {
			continue
		}

		// Following character has to be ":"
		if i >= len(buf)-2 || buf[i+1] != ':' {
			correctByteCount = 0
			continue
		}

		// we found the whole param!
		valueStartIdx = i + 2
		break
	}

	if correctByteCount < len(paramName) {
		return nil, errors.New("param \"" + string(paramName) + "\" not found in buffer")
	}
	if valueStartIdx >= len(buf)-1 {
		return nil, errors.New("nothing in buffer after \"" + string(paramName) + ":\"")
	}

	// Read all whitespace
	for {
		if buf[valueStartIdx] != ' ' {
			break
		}
		valueStartIdx++
	}

	// Read until CRLF
	return readUntilCrlf(buf[valueStartIdx:])
}

/* Reads from start of slice until CRLF. If no CRLF is found, it will return an error instead of the value so far. */
func readUntilCrlf(buf []byte) ([]byte, error) {
	lastTokenIdx := -1337
	for i, b := range buf {
		if b == '\r' {
			lastTokenIdx = i
		} else if b == '\n' {
			if lastTokenIdx == i-1 {
				return buf[:lastTokenIdx], nil
			}
		}
	}

	// we never found a valid CRLF
	return nil, errors.New("no CRLF found")
}

func isValidUpgradeRequest(buf []byte) (bool, error) {
	// TODO: This doesn't verify a valid HTTP verb at all

	// _, err := GetHttpParam(buf, "Host")
	// if err != nil {
	// 	return false, err
	// }

	httpConnection, err := getHttpParam(buf, "Connection")
	if err != nil || (string(httpConnection) != "Upgrade" && string(httpConnection) != "upgrade") {
		return false, errors.New("invalid or nonexistent \"Connection\" param")
	}

	httpUpgrade, err := getHttpParam(buf, "Upgrade")
	if err != nil || string(httpUpgrade) != "websocket" {
		return false, errors.New("invalid or nonexistent \"Upgrade\" param")
	}

	httpWebSocketVersion, err := getHttpParam(buf, "Sec-WebSocket-Version")
	if err != nil || string(httpWebSocketVersion) != "13" {
		return false, errors.New("invalid or nonexistent \"Sec-WebSocket-Version\" param")
	}

	_, err = getHttpParam(buf, "Sec-WebSocket-Key")
	if err != nil {
		return false, errors.New("invalid or nonexistent \"Sec-WebSocket-Key\" param")
	}

	return true, nil
}

func sendBadRequestResponse(conn *net.Conn) (int, error) {
	return (*conn).Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
}
