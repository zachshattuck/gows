package gows

import (
	"encoding/hex"
	"reflect"
	"testing"
)

/* "Hello, world!"
0x818d0e21bf6d4644d301610d9f1a6153d3092f
{1 0 0 0 1 1 [14 33 191 109] [72 101 108 108 111 44 32 119 111 114 108 100 33]}
*/

/* 0x01020304
0x82843283033333810037
{1 0 0 0 1 2 [50 131 3 51] [1 2 3 4]}
*/

func TestDeserializeWebSocketFrameMaskedText(t *testing.T) {
	got, err := DeserializeWebSocketFrame([]byte{0x81, 0x8d, 0x0e, 0x21, 0xbf, 0x6d, 0x46, 0x44, 0xd3, 0x01, 0x61, 0x0d, 0x9f, 0x1a, 0x61, 0x53, 0xd3, 0x09, 0x2f})
	want := WebSocketFrame{
		Payload:  []byte("Hello, world!"),
		Fin:      1,
		Rsv1:     0,
		Rsv2:     0,
		Rsv3:     0,
		IsMasked: 1,
		MaskKey:  [4]byte{14, 33, 191, 109},
		Opcode:   1}

	if err != nil {
		t.Errorf("got error: %s", err)
	}

	if got.Fin != want.Fin {
		t.Errorf("[fin] got %d, wanted %d", got.Fin, want.Fin)
	}
	if got.Rsv1 != want.Rsv1 {
		t.Errorf("[rsv1] got %d, wanted %d", got.Rsv1, want.Rsv1)
	}
	if got.Rsv2 != want.Rsv2 {
		t.Errorf("[rsv2] got %d, wanted %d", got.Rsv2, want.Rsv2)
	}
	if got.Rsv3 != want.Rsv3 {
		t.Errorf("[rsv3] got %d, wanted %d", got.Rsv3, want.Rsv3)
	}
	if got.Opcode != want.Opcode {
		t.Errorf("[opcode] got %d, wanted %d", got.Opcode, want.Opcode)
	}
	if got.IsMasked != want.IsMasked {
		t.Errorf("[isMasked] got %d, wanted %d", got.IsMasked, want.IsMasked)
	}

	if string(got.Payload) != string(want.Payload) {
		t.Errorf("[payload]\ngot:  %q\nwant: %q", string(got.Payload), string(want.Payload))
	}

	if reflect.DeepEqual(got.MaskKey, want.MaskKey) == false {
		t.Errorf("[maskKey]\ngot:  0x%s\nwant: 0x%s", hex.EncodeToString(got.MaskKey[:]), hex.EncodeToString(want.MaskKey[:]))
	}

}

func TestSerializeWebSocketFrameMaskedText(t *testing.T) {
	got := SerializeWebSocketFrame(WebSocketFrame{
		Payload:  []byte("Hello, world!"),
		Fin:      1,
		Rsv1:     0,
		Rsv2:     0,
		Rsv3:     0,
		IsMasked: 1,
		MaskKey:  [4]byte{14, 33, 191, 109},
		Opcode:   1})
	want := []byte{0x81, 0x8d, 0x0e, 0x21, 0xbf, 0x6d, 0x46, 0x44, 0xd3, 0x01, 0x61, 0x0d, 0x9f, 0x1a, 0x61, 0x53, 0xd3, 0x09, 0x2f}

	if len(got) != len(want) {
		t.Errorf("got length %d, wanted %d", len(got), len(want))
	}

	if reflect.DeepEqual(got, want) == false {
		t.Errorf("\ngot:  0x%s\nwant: 0x%s", hex.EncodeToString(got), hex.EncodeToString(want))
	}
}

func TestSerializeWebSocketFrameMaskedBinary(t *testing.T) {
	got := SerializeWebSocketFrame(WebSocketFrame{
		Payload:  []byte{1, 2, 3, 4},
		Fin:      1,
		Rsv1:     0,
		Rsv2:     0,
		Rsv3:     0,
		IsMasked: 1,
		MaskKey:  [4]byte{50, 131, 3, 51},
		Opcode:   2})
	want := []byte{0x82, 0x84, 0x32, 0x83, 0x03, 0x33, 0x33, 0x81, 0x00, 0x37}

	if len(got) != len(want) {
		t.Errorf("got length %d, wanted %d", len(got), len(want))
	}

	if reflect.DeepEqual(got, want) == false {
		t.Errorf("\ngot:  0x%s\nwant: 0x%s", hex.EncodeToString(got), hex.EncodeToString(want))
	}
}
