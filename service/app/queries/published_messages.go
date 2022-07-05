package queries

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type PublishedMessages struct {
	// Only messages with a sequence greater or equal to the start sequence are returned.
	StartSeq message.Sequence
}

type PublishedMessagesHandler struct {
	repository FeedRepository
	feed       refs.Feed
}

func NewPublishedMessagesHandler(repository FeedRepository, local identity.Public) (*PublishedMessagesHandler, error) {
	localRef, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a public identity")
	}

	return &PublishedMessagesHandler{
		repository: repository,
		feed:       localRef.MainFeed(),
	}, nil
}

func (h *PublishedMessagesHandler) Handle(query PublishedMessages) ([]message.Message, error) {
	if query.StartSeq.IsZero() {
		return nil, errors.New("zero value of sequence")
	}

	msgs, err := h.repository.GetMessages(h.feed, &query.StartSeq, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting messages")
	}

	return msgs, nil
}
