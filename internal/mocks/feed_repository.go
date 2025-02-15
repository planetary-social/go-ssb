package mocks

import (
	"fmt"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type FeedRepositoryMockGetMessagesCall struct {
	Id    refs.Feed
	Seq   *message.Sequence
	Limit *int
}

type FeedRepositoryMockGetMessageCall struct {
	Feed refs.Feed
	Seq  message.Sequence
}

type FeedRepositoryMockUpdateFeedCall struct {
	Feed refs.Feed
}

type FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall struct {
	Feed              refs.Feed
	MessagesToPersist []refs.Message
}

type FeedRepositoryMockRemoveMessagesAtOrAboveSequenceCall struct {
	Feed refs.Feed
	Seq  message.Sequence
}

type FeedRepositoryMock struct {
	getMessagesCalls       []FeedRepositoryMockGetMessagesCall
	GetMessagesReturnValue []message.Message
	GetMessagesReturnErr   error

	getMessageCalls        []FeedRepositoryMockGetMessageCall
	getMessageReturnValues map[string]message.Message

	updateFeedCalls                   []FeedRepositoryMockUpdateFeedCall
	updateFeedIgnoringReceiveLogCalls []FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall

	GetFeedCalls       []refs.Feed
	GetFeedReturnValue *feeds.Feed

	CountReturnValue int

	lock                                 sync.Mutex
	RemoveMessagesAtOrAboveSequenceCalls []FeedRepositoryMockRemoveMessagesAtOrAboveSequenceCall
}

func NewFeedRepositoryMock() *FeedRepositoryMock {
	return &FeedRepositoryMock{
		getMessageReturnValues: make(map[string]message.Message),
	}
}

func (m *FeedRepositoryMock) GetSequence(ref refs.Feed) (message.Sequence, error) {
	return message.Sequence{}, common.ErrFeedNotFound
}

func (m *FeedRepositoryMock) UpdateFeed(ref refs.Feed, f commands.UpdateFeedFn) error {
	call := FeedRepositoryMockUpdateFeedCall{
		Feed: ref,
	}
	m.updateFeedCalls = append(m.updateFeedCalls, call)
	return nil
}

func (m *FeedRepositoryMock) UpdateFeedIgnoringReceiveLog(ref refs.Feed, f commands.UpdateFeedFn) error {
	call := FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall{
		Feed: ref,
	}
	defer func() {
		m.updateFeedIgnoringReceiveLogCalls = append(m.updateFeedIgnoringReceiveLogCalls, call)
	}()

	feed := feeds.NewFeed(nil)
	if err := f(feed); err != nil {
		return errors.Wrap(err, "error")
	}
	messagesToPersist := feed.PopForPersisting()
	var messageRefs []refs.Message
	for _, msg := range messagesToPersist {
		messageRefs = append(messageRefs, msg.Message().Id())
	}
	call.MessagesToPersist = messageRefs
	return nil
}

func (m *FeedRepositoryMock) DeleteFeed(ref refs.Feed) error {
	return errors.New("not implemented")
}

func (m *FeedRepositoryMock) RemoveMessagesAtOrAboveSequence(ref refs.Feed, sequence message.Sequence) error {
	m.RemoveMessagesAtOrAboveSequenceCalls = append(m.RemoveMessagesAtOrAboveSequenceCalls, FeedRepositoryMockRemoveMessagesAtOrAboveSequenceCall{
		Feed: ref,
		Seq:  sequence,
	})
	return nil
}

func (m *FeedRepositoryMock) GetMessages(id refs.Feed, seq *message.Sequence, limit *int) ([]message.Message, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.getMessagesCalls = append(m.getMessagesCalls, FeedRepositoryMockGetMessagesCall{Id: id, Seq: seq, Limit: limit})
	return m.GetMessagesReturnValue, m.GetMessagesReturnErr
}

func (m *FeedRepositoryMock) MockGetMessage(msg message.Message) {
	m.getMessageReturnValues[fmt.Sprintf("%s-%d", msg.Feed().String(), msg.Sequence().Int())] = msg
}

func (m *FeedRepositoryMock) GetMessage(feed refs.Feed, sequence message.Sequence) (message.Message, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.getMessageCalls = append(m.getMessageCalls, FeedRepositoryMockGetMessageCall{Feed: feed, Seq: sequence})

	msg, ok := m.getMessageReturnValues[fmt.Sprintf("%s-%d", feed.String(), sequence.Int())]
	if !ok {
		return message.Message{}, errors.New("message not mocked")
	}
	return msg, nil
}

func (m *FeedRepositoryMock) GetFeed(ref refs.Feed) (*feeds.Feed, error) {
	m.GetFeedCalls = append(m.GetFeedCalls, ref)
	if m.GetFeedReturnValue == nil {
		return nil, errors.Wrap(common.ErrFeedNotFound, "feed not mocked")
	}
	return m.GetFeedReturnValue, nil
}

func (m *FeedRepositoryMock) Count() (int, error) {
	return m.CountReturnValue, nil
}

func (m *FeedRepositoryMock) GetMessagesCalls() []FeedRepositoryMockGetMessagesCall {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]FeedRepositoryMockGetMessagesCall, len(m.getMessagesCalls))
	copy(tmp, m.getMessagesCalls)
	return tmp
}

func (m *FeedRepositoryMock) GetMessageCalls() []FeedRepositoryMockGetMessageCall {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]FeedRepositoryMockGetMessageCall, len(m.getMessageCalls))
	copy(tmp, m.getMessageCalls)
	return tmp
}

func (m *FeedRepositoryMock) UpdateFeedIgnoringReceiveLogCalls() []FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]FeedRepositoryMockUpdateFeedIgnoringReceiveLogCall, len(m.updateFeedIgnoringReceiveLogCalls))
	copy(tmp, m.updateFeedIgnoringReceiveLogCalls)
	return tmp
}

func (m *FeedRepositoryMock) UpdateFeedCalls() []FeedRepositoryMockUpdateFeedCall {
	m.lock.Lock()
	defer m.lock.Unlock()

	tmp := make([]FeedRepositoryMockUpdateFeedCall, len(m.updateFeedCalls))
	copy(tmp, m.updateFeedCalls)
	return tmp
}
