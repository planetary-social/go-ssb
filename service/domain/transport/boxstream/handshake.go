package boxstream

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"go.cryptoscope.co/secretstream/secrethandshake"
)

// Handshaker performs the Secret Handshake using the provided ReadWriteCloser.
type Handshaker struct {
	local      identity.Private
	networkKey NetworkKey
}

// NewHandshaker creates a new handshaker which uses the provided local private
// identity when performing secret handshakes.
func NewHandshaker(local identity.Private, networkKey NetworkKey) (Handshaker, error) {
	if local.IsZero() {
		return Handshaker{}, errors.New("zero value of private identity")
	}

	if networkKey.IsZero() {
		return Handshaker{}, errors.New("zero value of network key")
	}

	return Handshaker{
		local:      local,
		networkKey: networkKey,
	}, nil
}

// OpenClientStream opens a client stream using the provided identity of the
// remote peer and the provided ReadWriteCloser. This should be used when
// initiating a connection with a remote peer.
func (h Handshaker) OpenClientStream(rw io.ReadWriteCloser, remote identity.Public) (*Stream, error) {
	if remote.IsZero() {
		return nil, errors.New("zero value of remote identity")
	}

	state, err := secrethandshake.NewClientState(h.networkKey.Bytes(), h.localKeypair(), remote.PublicKey())
	if err != nil {
		return nil, errors.Wrap(err, "could not create client state")
	}

	err = secrethandshake.Client(state, rw)
	if err != nil {
		return nil, errors.Wrap(err, "could not perform the client handshake")
	}

	return h.createStream(rw, state)
}

// OpenServerStream opens a server stream using the provided ReadWriteCloser.
// This should be used when handling incoming connections which were initiated
// by the other party.
func (h Handshaker) OpenServerStream(rw io.ReadWriteCloser) (*Stream, error) {
	state, err := secrethandshake.NewServerState(h.networkKey.Bytes(), h.localKeypair())
	if err != nil {
		return nil, errors.Wrap(err, "could not create client state")
	}

	err = secrethandshake.Server(state, rw)
	if err != nil {
		return nil, errors.Wrap(err, "could not perform the client handshake")
	}

	return h.createStream(rw, state)
}

func (h Handshaker) createStream(rw io.ReadWriteCloser, state *secrethandshake.State) (*Stream, error) {
	result, err := h.toResult(state)
	if err != nil {
		return nil, errors.Wrap(err, "converting to handshake result failed")
	}

	return NewStream(rw, result)
}

func (h Handshaker) localKeypair() secrethandshake.EdKeyPair {
	return secrethandshake.EdKeyPair{
		Public: h.local.Public().PublicKey(),
		Secret: h.local.PrivateKey(),
	}
}

func (h Handshaker) toResult(s *secrethandshake.State) (HandshakeResult, error) {
	remote, err := identity.NewPublicFromBytes(s.Remote())
	if err != nil {
		return HandshakeResult{}, errors.Wrap(err, "could not create a public")
	}

	var result HandshakeResult
	result.Remote = remote
	result.WriteSecret, result.WriteNonce = s.GetBoxstreamEncKeys()
	result.ReadSecret, result.ReadNonce = s.GetBoxstreamDecKeys()
	return result, nil
}
