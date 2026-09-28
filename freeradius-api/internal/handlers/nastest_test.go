package handlers

import (
	"context"
	"net"
	"testing"
	"time"

	"layeh.com/radius"
)

// TestCoAExchangeRoundTrip starts a fake NAS that ACKs Disconnect/CoA and
// checks that a signed request is accepted and the ACK is parsed.
func TestCoAExchangeRoundTrip(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot bind udp: %v", err)
	}
	secret := []byte("secret123")
	server := &radius.PacketServer{
		SecretSource: radius.StaticSecretSource(secret),
		Handler: radius.HandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
			var code radius.Code
			switch r.Code {
			case radius.CodeDisconnectRequest:
				code = radius.CodeDisconnectACK
			case radius.CodeCoARequest:
				code = radius.CodeCoAACK
			default:
				code = radius.CodeAccessReject
			}
			_ = w.Write(r.Response(code))
		}),
	}
	go func() { _ = server.Serve(conn) }()
	defer server.Shutdown(context.Background())

	for _, tc := range []struct {
		req  radius.Code
		want radius.Code
	}{
		{radius.CodeDisconnectRequest, radius.CodeDisconnectACK},
		{radius.CodeCoARequest, radius.CodeCoAACK},
	} {
		packet := radius.New(tc.req, secret)
		if err := signMessageAuthenticator(packet, string(secret)); err != nil {
			t.Fatalf("sign: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := radius.Exchange(ctx, packet, conn.LocalAddr().String())
		cancel()
		if err != nil {
			t.Fatalf("%s exchange: %v", tc.req, err)
		}
		if resp.Code != tc.want {
			t.Errorf("%s: got %s, want %s", tc.req, resp.Code, tc.want)
		}
	}
}
