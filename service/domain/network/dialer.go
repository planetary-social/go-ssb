package network

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

const (
	dialTimeout = 15 * time.Second
)

type ClientPeerInitializer interface {
	// InitializeClientPeer initializes outgoing connections by performing a handshake and establishing an RPC
	// connection using the provided ReadWriteCloser. Context is used as the RPC connection context.
	InitializeClientPeer(ctx context.Context, rwc io.ReadWriteCloser, remote identity.Public) (transport.Peer, error)
}

type Dialer struct {
	initializer ClientPeerInitializer
	logger      logging.Logger
}

func NewDialer(initializer ClientPeerInitializer, logger logging.Logger) (*Dialer, error) {
	return &Dialer{
		initializer: initializer,
		logger:      logger.New("transport"),
	}, nil
}

func (d Dialer) DialWithInitializer(ctx context.Context, initializer ClientPeerInitializer, remote identity.Public, addr Address) (transport.Peer, error) {
	dialer := net.Dialer{
		Timeout: dialTimeout,
	}

	conn, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not dial")
	}

	peer, err := initializer.InitializeClientPeer(ctx, conn, remote)
	if err != nil {
		return transport.Peer{}, errors.Wrap(err, "could not initialize a client peer")
	}

	return peer, nil
}

func (d Dialer) Dial(ctx context.Context, remote identity.Public, address Address) (transport.Peer, error) {
	return d.DialWithInitializer(ctx, d.initializer, remote, address)
}
