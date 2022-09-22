package ebt

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}

type MessageWriter interface {
	WriteMessage(msg message.Message) error
}

type MessageStreamer interface {
	Handle(ctx context.Context, id refs.Feed, seq *message.Sequence, messageWriter MessageWriter)
}

type Stream interface {
	IncomingMessages(ctx context.Context) <-chan IncomingMessage
	SendNote(note messages.EbtReplicateNote)
	SendMessage(msg *message.Message)
}

type IncomingMessage struct {
	notes *messages.EbtReplicateNotes
	msg   *message.RawMessage
	err   error
}

func NewIncomingMessageWithNote(notes messages.EbtReplicateNotes) IncomingMessage {
	return IncomingMessage{
		notes: &notes,
	}
}

func NewIncomingMessageWithMesage(msg message.RawMessage) IncomingMessage {
	return IncomingMessage{
		msg: &msg,
	}
}

func NewIncomingMessageWithErr(err error) IncomingMessage {
	return IncomingMessage{
		err: err,
	}
}

func (i IncomingMessage) Notes() (messages.EbtReplicateNotes, bool) {
	if i.notes != nil {
		return *i.notes, true
	}
	return messages.EbtReplicateNotes{}, false
}

func (i IncomingMessage) Msg() (message.RawMessage, bool) {
	if i.msg != nil {
		return *i.msg, true
	}
	return message.RawMessage{}, false
}

func (i IncomingMessage) Err() error {
	return i.err
}

type SessionRunner struct {
	logger            logging.Logger
	rawMessageHandler RawMessageHandler
	streamer          MessageStreamer
}

func NewSessionRunner(
	logger logging.Logger,
	rawMessageHandler RawMessageHandler,
	streamer MessageStreamer,
) *SessionRunner {
	return &SessionRunner{
		logger:            logger,
		rawMessageHandler: rawMessageHandler,
		streamer:          streamer,
	}
}

func (s *SessionRunner) HandleStream(ctx context.Context, stream Stream) error {
	session := NewSession(stream, s.logger, s.rawMessageHandler, s.streamer)
	go session.handleIncomingMessages(ctx)

	// todo send

	return nil
}

type Session struct {
	stream Stream

	remoteNotes map[string]messages.EbtReplicateNote
	lock        sync.Mutex // guards remoteNotes

	logger            logging.Logger
	rawMessageHandler RawMessageHandler
	streamer          MessageStreamer
}

func NewSession(
	stream Stream,
	logger logging.Logger,
	rawMessageHandler RawMessageHandler,
	streamer MessageStreamer,
) *Session {
	return &Session{
		stream: stream,

		logger:            logger.New("session"),
		rawMessageHandler: rawMessageHandler,
		streamer:          streamer,
	}
}

func (s *Session) handleIncomingMessages(ctx context.Context) {
	for incoming := range s.stream.IncomingMessages(ctx) {
		if err := s.handleIncomingMessage(ctx, incoming); err != nil {
			s.logger.WithError(err).Debug("error processing incoming message")
		}
	}
}

func (s *Session) handleIncomingMessage(ctx context.Context, incoming IncomingMessage) error {
	if err := incoming.Err(); err != nil {
		return errors.Wrap(err, "error receiving messages")
	}

	notes, ok := incoming.Notes()
	if ok {
		return s.handleIncomingNotes(ctx, notes)
	}

	msg, ok := incoming.Msg()
	if ok {
		if err := s.rawMessageHandler.Handle(msg); err != nil {
			return errors.Wrap(err, "error handling a message")
		}
		return nil
	}

	return errors.New("logic error")
}

func (s *Session) handleIncomingNotes(ctx context.Context, notes messages.EbtReplicateNotes) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, note := range notes.Notes() {
		s.remoteNotes[note.Ref().String()] = note

		seq, err := s.parseSeq(note.Sequence())
		if err != nil {
			return errors.Wrap(err, "error parsing sequence")
		}

		// todo kill old streams
		s.streamer.Handle(ctx, note.Ref(), seq, NewStreamMessageWriter(s.stream))
	}

	return nil
}

func (s *Session) parseSeq(seq int) (*message.Sequence, error) {
	if seq <= 0 {
		return nil, nil
	}
	sequence, err := message.NewSequence(seq)
	if err != nil {
		return nil, errors.Wrap(err, "new sequence error")
	}
	return &sequence, nil
}

type StreamMessageWriter struct {
	stream Stream
}

func NewStreamMessageWriter(stream Stream) *StreamMessageWriter {
	return &StreamMessageWriter{
		stream: stream,
	}
}

func (s StreamMessageWriter) WriteMessage(msg message.Message) error {
	s.stream.SendMessage(&msg)
	return nil
}
