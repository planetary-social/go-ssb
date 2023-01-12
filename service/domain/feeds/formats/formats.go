package formats

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type Marshaler interface {
	Marshal(content content.KnownMessageContent) (message.RawMessageContent, error)
	Unmarshal(b message.RawMessageContent) (content.KnownMessageContent, error)
}

type RawMessageIdentifier struct {
	formats []feeds.FeedFormat
}

func NewRawMessageIdentifier(formats []feeds.FeedFormat) *RawMessageIdentifier {
	return &RawMessageIdentifier{
		formats: formats,
	}
}

func (i RawMessageIdentifier) VerifyRawMessage(raw message.RawMessage) (message.Message, error) {
	var result error

	for _, format := range i.formats {
		msg, err := format.Verify(raw)
		if err == nil {
			return msg, nil
		}

		result = multierror.Append(result, err)
	}

	return message.Message{}, errors.Wrap(result, "unknown message")
}

func (i RawMessageIdentifier) LoadRawMessage(raw message.VerifiedRawMessage) (message.MessageWithoutId, error) {
	var result error

	for _, format := range i.formats {
		msg, err := format.Load(raw)
		if err == nil {
			return msg, nil
		}

		result = multierror.Append(result, err)
	}

	return message.MessageWithoutId{}, errors.Wrap(result, "unknown message")
}

// MessageHMACLength is implied to be constant due to an assumption that this
// key is used as an HMAC key when calling libsodium's functions.
const MessageHMACLength = 32

// MessageHMAC is mainly used for test networks. It is applied to messages when
// creating them to make them incompatible with the main Secure Scuttlebutt
// network. MessageHMAC is not documented in the Protocol Guide as by default
// it is not applied, it is more of a convention.
//
// See https://github.com/ssb-js/ssb-validate#state--validateappendstate-hmac_key-msg.
type MessageHMAC struct {
	b []byte
}

// NewMessageHMAC creates a message HMAC from the provided slice of bytes. The
// slice must have a length of MessageHMACLength or 0. Passing slice of length
// 0 is equivalent to calling NewDefaultMessageHMAC which should be preferred
// for redability.
func NewMessageHMAC(b []byte) (MessageHMAC, error) {
	if len(b) == 0 {
		return NewDefaultMessageHMAC(), nil
	}

	if len(b) != MessageHMACLength {
		return MessageHMAC{}, fmt.Errorf("invalid message HMAC length, must be '%d'", MessageHMACLength)
	}

	buf := make([]byte, MessageHMACLength)
	copy(buf, b)
	return MessageHMAC{buf}, nil
}

func MustNewMessageHMAC(b []byte) MessageHMAC {
	v, err := NewMessageHMAC(b)
	if err != nil {
		panic(err)
	}
	return v
}

// NewDefaultMessageHMAC returns a MessageHMAC used by the main Secure
// Secuttlebutt network. This value effectively means that message HMAC should
// not be applied to messages.
func NewDefaultMessageHMAC() MessageHMAC {
	return MessageHMAC{nil}
}

// Bytes returns the slice of length MessageHMACLength or nil if this is the
// default message HMAC.
func (k MessageHMAC) Bytes() []byte {
	if k.IsZero() {
		return nil
	}

	tmp := make([]byte, len(k.b))
	copy(tmp, k.b)
	return tmp
}

func (k MessageHMAC) IsZero() bool {
	return len(k.b) == 0
}
