package message

import (
	"fmt"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type UnsignedMessage struct {
	baseMessageFields
	content RawContent
}

func NewUnsignedMessage(
	previous *refs.Message,
	sequence Sequence,
	author refs.Identity,
	feed refs.Feed,
	timestamp time.Time,
	content RawContent,
) (UnsignedMessage, error) {
	fields, err := newBaseMessageFields(previous, sequence, author, feed, timestamp)
	if err != nil {
		return UnsignedMessage{}, errors.Wrap(err, "could not create base message fields")
	}

	if content.IsZero() {
		return UnsignedMessage{}, errors.Wrap(err, "zero value of content")
	}

	return UnsignedMessage{
		baseMessageFields: fields,
		content:           content,
	}, nil
}

func (m UnsignedMessage) Content() RawContent {
	return m.content
}

type Message struct {
	baseMessageFields
	content Content

	id  refs.Message
	raw RawMessage
}

func NewMessage(
	id refs.Message,
	previous *refs.Message,
	sequence Sequence,
	author refs.Identity,
	feed refs.Feed,
	timestamp time.Time,
	content Content,
	raw RawMessage,
) (Message, error) {
	fields, err := newBaseMessageFields(previous, sequence, author, feed, timestamp)
	if err != nil {
		return Message{}, errors.Wrap(err, "could not create base message fields")
	}

	if id.IsZero() {
		return Message{}, errors.New("zero value of id")
	}

	if content.IsZero() {
		return Message{}, errors.New("zero value of content")
	}

	if raw.IsZero() {
		return Message{}, errors.New("zero value of raw message")
	}

	return Message{
		id:                id,
		baseMessageFields: fields,
		content:           content,
		raw:               raw,
	}, nil
}

func MustNewMessage(
	id refs.Message,
	previous *refs.Message,
	sequence Sequence,
	author refs.Identity,
	feed refs.Feed,
	timestamp time.Time,
	content Content,
	raw RawMessage,
) Message {
	msg, err := NewMessage(id, previous, sequence, author, feed, timestamp, content, raw)
	if err != nil {
		panic(err)
	}
	return msg
}

func NewMessageFromMessageWithoutId(id refs.Message, msgWithoutId MessageWithoutId) (Message, error) {
	return NewMessage(
		id,
		msgWithoutId.Previous(),
		msgWithoutId.Sequence(),
		msgWithoutId.Author(),
		msgWithoutId.Feed(),
		msgWithoutId.Timestamp(),
		msgWithoutId.Content(),
		msgWithoutId.Raw(),
	)
}

func (m Message) Id() refs.Message {
	return m.id
}

func (m Message) Content() Content {
	return m.content
}

func (m Message) Raw() RawMessage {
	return m.raw
}

func (m Message) IsRootMessage() bool {
	return m.sequence.IsFirst()
}

func (m Message) ComesDirectlyBefore(o Message) bool {
	if o.previous == nil {
		return false
	}
	return m.sequence.ComesDirectlyBefore(o.sequence) && o.previous.Equal(m.id)
}

func (m Message) IsZero() bool {
	return m.id.IsZero()
}

func (m Message) String() string {
	return fmt.Sprintf("id=%s feed=%s sequence=%d previous=%s", m.id.String(), m.feed.String(), m.sequence.Int(), m.previous)
}

type baseMessageFields struct {
	previous  *refs.Message
	sequence  Sequence
	author    refs.Identity
	feed      refs.Feed
	timestamp time.Time
}

func newBaseMessageFields(
	previous *refs.Message,
	sequence Sequence,
	author refs.Identity,
	feed refs.Feed,
	timestamp time.Time,
) (baseMessageFields, error) {
	if previous != nil && previous.IsZero() {
		return baseMessageFields{}, errors.New("zero value of previous")
	}

	if previous != nil && sequence.IsFirst() {
		return baseMessageFields{}, errors.New("this message has a previous message so it can't be first")
	}

	if previous == nil && !sequence.IsFirst() {
		return baseMessageFields{}, errors.New("this message doesn't have a previous message so it must be first")
	}

	if author.IsZero() {
		return baseMessageFields{}, errors.New("zero value of author")
	}

	if feed.IsZero() {
		return baseMessageFields{}, errors.New("zero value of feed")
	}

	// there is no way to validate the timestamp as it can be set to anything

	return baseMessageFields{
		previous:  previous,
		sequence:  sequence,
		author:    author,
		feed:      feed,
		timestamp: timestamp,
	}, nil
}

func (k baseMessageFields) Previous() *refs.Message {
	return k.previous
}

func (k baseMessageFields) Sequence() Sequence {
	return k.sequence
}

func (k baseMessageFields) Author() refs.Identity {
	return k.author
}

func (k baseMessageFields) Feed() refs.Feed {
	return k.feed
}

func (k baseMessageFields) Timestamp() time.Time {
	return k.timestamp
}
