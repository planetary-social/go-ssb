package feeds_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/stretchr/testify/require"
)

func TestNewContactFromHistory(t *testing.T) {
	target := fixtures.SomeRefIdentity()

	t.Run("following_true", func(t *testing.T) {
		c, err := feeds.NewContactFromHistory(target, true, false)
		require.NoError(t, err)

		require.Equal(t, target, c.Target())
		require.True(t, c.Following())
		require.False(t, c.Blocking())
	})

	t.Run("following_false", func(t *testing.T) {
		c, err := feeds.NewContactFromHistory(target, false, true)
		require.NoError(t, err)

		require.Equal(t, target, c.Target())
		require.False(t, c.Following())
		require.True(t, c.Blocking())
	})
}

func TestContact_Update(t *testing.T) {
	testCases := []struct {
		Name      string
		Actions   []content.ContactAction
		Following bool
		Blocking  bool
	}{
		{
			Name: "follow",
			Actions: []content.ContactAction{
				content.ContactActionFollow,
			},
			Following: true,
			Blocking:  false,
		},
		{
			Name: "unfollow",
			Actions: []content.ContactAction{
				content.ContactActionUnfollow,
			},
			Following: false,
			Blocking:  false,
		},
		{
			Name: "block",
			Actions: []content.ContactAction{
				content.ContactActionBlock,
			},
			Following: false,
			Blocking:  true,
		},
		{
			Name: "unblock",
			Actions: []content.ContactAction{
				content.ContactActionUnblock,
			},
			Following: false,
			Blocking:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			contact, err := feeds.NewContact(fixtures.SomeRefIdentity())
			require.NoError(t, err)

			err = contact.Update(content.MustNewContactActions(testCase.Actions))
			require.NoError(t, err)

			require.Equal(t, testCase.Following, contact.Following())
			require.Equal(t, testCase.Blocking, contact.Blocking())
		})
	}
}

func TestContact_UpdateCorrectlyAppliesAllActions(t *testing.T) {
	actions, err := content.NewContactActions(
		[]content.ContactAction{
			content.ContactActionUnfollow,
			content.ContactActionUnblock,
		},
	)
	require.NoError(t, err)

	contact, err := feeds.NewContactFromHistory(fixtures.SomeRefIdentity(), true, true)
	require.NoError(t, err)

	require.True(t, contact.Following())
	require.True(t, contact.Blocking())

	err = contact.Update(actions)
	require.NoError(t, err)

	require.False(t, contact.Following())
	require.False(t, contact.Blocking())
}
