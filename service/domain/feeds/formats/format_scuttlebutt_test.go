package formats_test

import (
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestScuttlebutt_MessageCanBeSignedAndThenVerified(t *testing.T) {
	f := newScuttlebuttFormat(t, formats.NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message.NewUnsignedMessage(
		nil,
		message.NewFirstSequence(),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msgFromSign, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msgFromSign.Raw())
	require.NoError(t, err)

	require.Equal(t, msgFromSign, msgFromVerify)
}

func TestScuttlebutt_MessageCanBeSignedAndThenVerifiedIfItReferencesAPreviousMessage(t *testing.T) {
	f := newScuttlebuttFormat(t, formats.NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	previous := fixtures.SomeRefMessage()

	unsignedMessage, err := message.NewUnsignedMessage(
		&previous,
		message.MustNewSequence(2),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msgFromSign, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msgFromSign.Raw())
	require.NoError(t, err)

	require.Equal(t, msgFromSign, msgFromVerify)
}

func TestScuttlebutt_SettingHMACMakesMessagesIncompatibile(t *testing.T) {
	hmac, err := formats.NewMessageHMAC([]byte("somehmacthatislongenoughblablabl"))
	require.NoError(t, err)

	f := newScuttlebuttFormat(t, hmac)

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	unsignedMessage, err := message.NewUnsignedMessage(
		nil,
		message.NewFirstSequence(),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msg, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	_, err = f.Verify(msg.Raw())
	require.NoError(t, err)

	defaultFormat := newScuttlebuttFormat(t, formats.NewDefaultMessageHMAC())
	_, err = defaultFormat.Verify(msg.Raw())
	require.Contains(t, err.Error(), "invalid signature")
}

func TestScuttlebutt_LoadAndVerifyReturnIdenticalResults(t *testing.T) {
	f := newScuttlebuttFormat(t, formats.NewDefaultMessageHMAC())

	author, err := identity.NewPrivate()
	require.NoError(t, err)

	authorRef, err := refs.NewIdentityFromPublic(author.Public())
	require.NoError(t, err)

	previous := fixtures.SomeRefMessage()

	unsignedMessage, err := message.NewUnsignedMessage(
		&previous,
		message.MustNewSequence(2),
		authorRef,
		authorRef.MainFeed(),
		time.Now(),
		someContent(),
	)
	require.NoError(t, err)

	msg, err := f.Sign(unsignedMessage, author)
	require.NoError(t, err)

	verifiedRawMessage, err := message.NewVerifiedRawMessage(msg.Raw().Bytes())
	require.NoError(t, err)

	msgFromLoadWithoutId, err := f.Load(verifiedRawMessage)
	require.NoError(t, err)

	msgFromLoad, err := message.NewMessageFromMessageWithoutId(msg.Id(), msgFromLoadWithoutId)
	require.NoError(t, err)

	msgFromVerify, err := f.Verify(msg.Raw())
	require.NoError(t, err)

	require.Equal(t, msg, msgFromVerify)
	require.Equal(t, msgFromLoad, msgFromVerify)
}

func someContent() message.RawContent {
	return message.MustNewRawContent([]byte(`{"type": "something"}`))
}

func newScuttlebuttFormat(t *testing.T, hmac formats.MessageHMAC) *formats.Scuttlebutt {
	parser := newContentParserMock()
	return formats.NewScuttlebutt(parser, hmac)
}

type contentParserMock struct {
}

func newContentParserMock() *contentParserMock {
	return &contentParserMock{}
}

func (c contentParserMock) Parse(raw message.RawContent) (message.Content, error) {
	return message.NewContent(raw, nil, nil)
}
