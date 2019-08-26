package csms

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	sst "github.com/libp2p/go-libp2p-testing/suites/sec"
)

type TransportAdapter struct {
	mux *SSMuxer
}

func (sm *TransportAdapter) SecureInbound(ctx context.Context, insecure net.Conn) (sec.SecureConn, error) {
	sconn, _, err := sm.mux.SecureInbound(ctx, insecure)
	return sconn, err
}

func (sm *TransportAdapter) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	sconn, _, err := sm.mux.SecureOutbound(ctx, insecure, p)
	return sconn, err
}

func TestCommonProto(t *testing.T) {
	var at, bt SSMuxer
	atInsecure := insecure.New("peerA")
	btInsecure := insecure.New("peerB")
	at.AddTransport("/plaintext/1.0.0", atInsecure)
	bt.AddTransport("/plaintext/1.1.0", btInsecure)
	bt.AddTransport("/plaintext/1.0.0", btInsecure)
	sst.SubtestRW(t, &TransportAdapter{mux: &at}, &TransportAdapter{mux: &bt}, "peerA", "peerB")
}

func TestNoCommonProto(t *testing.T) {
	var at, bt SSMuxer
	atInsecure := insecure.New("peerA")
	btInsecure := insecure.New("peerB")

	at.AddTransport("/plaintext/1.0.0", atInsecure)
	bt.AddTransport("/plaintext/1.1.0", btInsecure)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, b := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer a.Close()
		_, _, err := at.SecureInbound(ctx, a)
		if err == nil {
			t.Fatal("conection should have failed")
		}
	}()

	go func() {
		defer wg.Done()
		defer b.Close()
		_, _, err := bt.SecureOutbound(ctx, b, "peerA")
		if err == nil {
			t.Fatal("connection should have failed")
		}
	}()
	wg.Wait()
}
