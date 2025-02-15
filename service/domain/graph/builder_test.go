package graph_test

import (
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestSocialGraphBuilder_FolloweesAreAddedToTheGraphUpToCertainDepth(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
		},
		b.String(): {
			feeds.MustNewContactFromHistory(b, c, true, false),
		},
	}

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)

	require.True(t, g.HasContact(local))
	require.True(t, g.HasContact(a))
	require.True(t, g.HasContact(b))
	require.False(t, g.HasContact(c))
}

func TestSocialGraphBuilder_SmallerNumberOfHopsTakesPriority(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
			feeds.MustNewContactFromHistory(local, b, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
		},
	}

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(1),
			},
		},
		g.Contacts(),
	)
}

func TestSocialGraphBuilder_EdgesWhichAreBlockedOrNotSetAsFollowingAreNotFollowed(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()
	e := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
			feeds.MustNewContactFromHistory(a, c, false, true),
			feeds.MustNewContactFromHistory(a, d, false, false),
			feeds.MustNewContactFromHistory(a, e, true, true),
		},
	}

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)
}

func TestContacts_LocalBlockingTakesPriorityAndAlwaysExcludesFeeds(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()
	e := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
			feeds.MustNewContactFromHistory(local, b, true, false),
			feeds.MustNewContactFromHistory(local, c, false, true),
			feeds.MustNewContactFromHistory(local, d, false, false),
			feeds.MustNewContactFromHistory(local, e, true, true),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
			feeds.MustNewContactFromHistory(a, c, true, false),
			feeds.MustNewContactFromHistory(a, d, true, false),
			feeds.MustNewContactFromHistory(a, e, true, false),
		},
	}

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(1),
			},
			{
				d,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)
}

func TestSocialGraphBuilder_BanningLocalFeedIsIgnored(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
		},
	}

	test.BanList.Ban(local.MainFeed())

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
		},
		g.Contacts(),
	)
}

func TestSocialGraphBuilder_BanListExcludesChildContacts(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
			feeds.MustNewContactFromHistory(local, b, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, c, true, false),
		},
		b.String(): {
			feeds.MustNewContactFromHistory(b, d, true, false),
			feeds.MustNewContactFromHistory(b, a, true, false),
		},
	}

	test.BanList.Ban(local.MainFeed())
	test.BanList.Ban(a.MainFeed())

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				b,
				graph.MustNewHops(1),
			},
			{
				d,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)
}

func TestSocialGraphBuilder_ContactsStorageIsNotNeedlesslyQueried(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
		},
		b.String(): {
			feeds.MustNewContactFromHistory(b, c, true, false),
		},
		c.String(): {
			feeds.MustNewContactFromHistory(c, d, true, false),
		},
	}

	g, err := test.Builder.Build()
	require.NoError(t, err)

	requireEqualContacts(t,
		[]graph.Contact{
			{
				local,
				graph.MustNewHops(0),
			},
			{
				a,
				graph.MustNewHops(1),
			},
			{
				b,
				graph.MustNewHops(2),
			},
		},
		g.Contacts(),
	)

	require.Equal(t,
		[]storageMockGetContactsCall{
			{
				Node: local,
			},
			{
				Node: a,
			},
		},
		test.ContactsStorage.GetContactsCalls,
	)
}

func TestSocialGraphBuilder_HasContactDoesNotNeedlesslyQueryContactsStorage(t *testing.T) {
	test := newTestSocialGraphBuilder(t)

	local := test.Local

	a := fixtures.SomeRefIdentity()
	b := fixtures.SomeRefIdentity()
	c := fixtures.SomeRefIdentity()
	d := fixtures.SomeRefIdentity()

	test.ContactsStorage.contacts = map[string][]*feeds.Contact{
		local.String(): {
			feeds.MustNewContactFromHistory(local, a, true, false),
		},
		a.String(): {
			feeds.MustNewContactFromHistory(a, b, true, false),
		},
		b.String(): {
			feeds.MustNewContactFromHistory(b, c, true, false),
		},
		c.String(): {
			feeds.MustNewContactFromHistory(c, d, true, false),
		},
	}

	ok, err := test.Builder.HasContact(a)
	require.NoError(t, err)
	require.True(t, ok)

	require.Equal(t,
		[]storageMockGetContactsCall{
			{
				Node: local,
			},
		},
		test.ContactsStorage.GetContactsCalls,
	)

	ok, err = test.Builder.HasContact(b)
	require.NoError(t, err)
	require.True(t, ok)

	require.Equal(t,
		[]storageMockGetContactsCall{
			{
				Node: local,
			},
			{
				Node: a,
			},
		},
		test.ContactsStorage.GetContactsCalls,
	)
}

func requireEqualContacts(t *testing.T, a []graph.Contact, b []graph.Contact) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Id.String() < a[j].Id.String()
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].Id.String() < b[j].Id.String()
	})
	require.Equal(t, a, b)
}

type TestSocialGraphBuilder struct {
	Builder *graph.SocialGraphBuilder

	ContactsStorage *storageMock
	BanList         *BanListMock

	Hops  graph.Hops
	Local refs.Identity
}

func newTestSocialGraphBuilder(t *testing.T) TestSocialGraphBuilder {
	contactsStorage := newStorageMock()
	banList := NewBanListMock()
	hops := graph.MustNewHops(2)
	local := fixtures.SomeRefIdentity()

	return TestSocialGraphBuilder{
		Builder:         graph.NewSocialGraphBuilder(contactsStorage, banList, hops, local),
		ContactsStorage: contactsStorage,
		BanList:         banList,

		Hops:  hops,
		Local: local,
	}

}

type storageMockGetContactsCall struct {
	Node refs.Identity
}

type storageMock struct {
	GetContactsCalls []storageMockGetContactsCall
	contacts         map[string][]*feeds.Contact
}

func newStorageMock() *storageMock {
	return &storageMock{}
}

func (s *storageMock) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	s.GetContactsCalls = append(s.GetContactsCalls, storageMockGetContactsCall{Node: node})
	return s.contacts[node.String()], nil
}

type BanListMock struct {
	bannedFeeds internal.Set[string]
}

func NewBanListMock() *BanListMock {
	return &BanListMock{
		bannedFeeds: internal.NewSet[string](),
	}
}

func (m BanListMock) Ban(feed refs.Feed) {
	m.bannedFeeds.Put(feed.String())
}

func (m BanListMock) ContainsFeed(feed refs.Feed) (bool, error) {
	return m.bannedFeeds.Contains(feed.String()), nil
}
