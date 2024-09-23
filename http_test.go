package gows

import "testing"

/* Example WebSocket upgrade request, ripped straight from my browser. */
var exampleHttpRequest = []byte("GET / HTTP/1.1\r\nHost: 127.0.0.1:8081\r\nConnection: Upgrade\r\nPragma: no-cache\r\nCache-Control: no-cache\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36\r\nUpgrade: websocket\r\nOrigin: http://localhost:8080\r\nSec-WebSocket-Version: 13\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Language: en-US,en;q=0.9\r\nSec-WebSocket-Key: D8KfDxohPIack4T9PAf3Ng==\r\nSec-WebSocket-Extensions: permessage-deflate; client_max_window_bits\r\n\r\n")

/* A longer WebSocket upgrade request, proxied by nginx. */
var exampleHttpRequest2 = []byte("GET /ws HTTP/1.1\r\nUpgrade: websocket\r\nConnection: upgrade\r\nHost: 127.0.0.1:8081\r\naccept-encoding: gzip, br\r\nX-Forwarded-For: 1.2.3.4\r\nCF-RAY: 8c3d6a50b90875c8-SEA\r\nX-Forwarded-Proto: https\r\nCF-Visitor: {\"scheme\":\"https\"}\r\nPragma: no-cache\r\nCache-Control: no-cache\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36\r\nOrigin: https://www.website.com\r\nSec-WebSocket-Version: 13\r\nAccept-Language: en-US,en;q=0.9\r\nSec-WebSocket-Key: ZFPbTE+Wekp3z+QNUR4R0Q==\r\nSec-WebSocket-Extensions: permessage-deflate; client_max_window_bits\r\nCF-Connecting-IP: 1.2.3.4\r\ncdn-loop: cloudflare; loops=1\r\nCF-IPCountry: US\r\n\r\n")

func TestGetHttpParamValidProperty(t *testing.T) {
	got, err := getHttpParam(exampleHttpRequest, "Host")
	want := "127.0.0.1:8081"

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestGetHttpParamInvalidProperty(t *testing.T) {
	got, err := getHttpParam(exampleHttpRequest, " super duper nonsensical parameter!!!!  89740r3n3yr0932")

	if err == nil {
		t.Errorf("did not error, got %q", got)
	}
}

func TestGetHttpParamDuplicateKeyword(t *testing.T) {
	got, err := getHttpParam(exampleHttpRequest, "Upgrade")
	want := "websocket"

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestGetHttpParamExtremelyLongParameter(t *testing.T) {
	got, err := getHttpParam(exampleHttpRequest, "sd09 fus8-d90f js09df mus90d8f mu09sd8fy um90s8d ynf098sd7f n908sd 7fn90s8d7fn 908sd7n f09s8df7n 908sd7f 098sd7nf 098sd7nf 098sd7nf 098sd7fn 098sd7nf 098sd7fn 098sd7f 098sdf7 n09s8df7n 098sdf7n 09s8df7 n09s8df 709s8df7 n09s8df7 098sdf7 098sdf yunoiusdf hlksjdfh klsjdfkjsdhfkj sdhjflksdjf lksdj f098sd7f 908sduf iujsdhf kjshdf kjysud9f8 7sd98f sdkjf h,sjdhf kjsdfy 98sdf iusdnf kjsdhf kiusdyf 98sdyfi uhsdifu ysd98f sd98f jsd98f jsd9f j9sd8f j9s8df hisudfh lkjsdhf8sdy f98sdhf iujsdhf iousdyuf 98sdhf oijsdhf likudsfyg s98ydfgisu hdfsiog hsdf98g y9sd8fgjh s9d8fg u9isd8fgy 0987sdfg yhioudsfhg oisudfgh o87sdfhg 9sdfgy h098sdfhg isdufhg 98sdfh g9087sdfhg iosdufhg osjkdfhg lkjdsfh giusdfug98dsfgu g9p8sdfjg ;lksdfj g")
	want := ""

	if err == nil || string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
func TestGetHttpParamSelf(t *testing.T) {
	got, err := getHttpParam(exampleHttpRequest, "GET / HTTP/1.1\r\nHost: 127.0.0.1:8081\r\nConnection: Upgrade\r\nPragma: no-cache\r\nCache-Control: no-cache\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36\r\nUpgrade: websocket\r\nOrigin: http://localhost:8080\r\nSec-WebSocket-Version: 13\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Language: en-US,en;q=0.9\r\nSec-WebSocket-Key: D8KfDxohPIack4T9PAf3Ng==\r\nSec-WebSocket-Extensions: permessage-deflate; client_max_window_bits\r\n\r\n")

	if err == nil {
		t.Errorf("did not error, got %q", got)
	}
}

func TestReadUntilCrlfHttpVerb(t *testing.T) {
	got, err := readUntilCrlf(exampleHttpRequest)
	want := "GET / HTTP/1.1"

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestReadUntilCrlfHttpParamValue(t *testing.T) {
	got, err := readUntilCrlf(exampleHttpRequest[22:])
	want := "127.0.0.1:8081"

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestReadUntilCrlfRandomNoCrlf(t *testing.T) {
	got, err := readUntilCrlf([]byte("This is a nice string and all, but it doesn't have a Crlf."))

	if err == nil {
		t.Errorf("did not error, got %q", got)
	}
}

func TestReadUntilCrlfRandomWithCrlf(t *testing.T) {
	got, err := readUntilCrlf([]byte("This is a nice string and all, AND it has a Crlf.\r\n"))
	want := "This is a nice string and all, AND it has a Crlf."

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestReadUntilCrlfCursed1(t *testing.T) {
	got, err := readUntilCrlf([]byte("\r \n\n\n\n \n\n\n\n\r\n\n\n\n\n\n\n"))
	want := "\r \n\n\n\n \n\n\n\n"

	if err != nil {
		t.Error("error:", err)
	}
	if string(got) != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
func TestReadUntilCrlfCursed2(t *testing.T) {
	got, err := readUntilCrlf([]byte("\r\r\r\r\r\r\r\r \nr\n\n\n \n\n\n\n\n\n\n\n\n\n\n"))

	if err == nil {
		t.Errorf("did not error, got %q", got)
	}
}

func TestIsValidUpgradeRequestBasicGood(t *testing.T) {
	got, err := isValidUpgradeRequest(exampleHttpRequest)

	if err != nil {
		t.Error("error:", err)
	}
	if got == false {
		t.Errorf("got invalid, expected valid")
	}
}

func TestIsValidUpgradeRequestLongGood(t *testing.T) {
	got, err := isValidUpgradeRequest(exampleHttpRequest2)

	if err != nil {
		t.Error("error:", err)
	}
	if got == false {
		t.Errorf("got invalid, expected valid")
	}
}
