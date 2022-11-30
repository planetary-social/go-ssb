// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	migrations2 "github.com/planetary-social/scuttlego/migrations"
	"github.com/planetary-social/scuttlego/service/adapters"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	ebt2 "github.com/planetary-social/scuttlego/service/adapters/ebt"
	"github.com/planetary-social/scuttlego/service/adapters/invites"
	"github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/planetary-social/scuttlego/service/adapters/mocks"
	"github.com/planetary-social/scuttlego/service/adapters/pubsub"
	"github.com/planetary-social/scuttlego/service/app"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain"
	replication2 "github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	invites2 "github.com/planetary-social/scuttlego/service/domain/invites"
	mocks2 "github.com/planetary-social/scuttlego/service/domain/mocks"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/network/local"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/gossip"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	transport2 "github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	network2 "github.com/planetary-social/scuttlego/service/ports/network"
	pubsub2 "github.com/planetary-social/scuttlego/service/ports/pubsub"
	rpc2 "github.com/planetary-social/scuttlego/service/ports/rpc"
	"go.etcd.io/bbolt"
)

// Injectors from wire.go:

func BuildTxTestAdapters(tx *bbolt.Tx) (TxTestAdapters, error) {
	messageContentMappings := transport.DefaultMappings()
	logger := fixtures.SomeLogger()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return TxTestAdapters{}, err
	}
	messageHMAC := formats.NewDefaultMessageHMAC()
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	private, err := identity.NewPrivate()
	if err != nil {
		return TxTestAdapters{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	graphHops := _wireHopsValue
	banListHasherMock := mocks.NewBanListHasherMock()
	banListRepository := bolt.NewBanListRepository(tx, banListHasherMock)
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops, banListRepository)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, banListRepository, scuttlebutt)
	currentTimeProviderMock := mocks.NewCurrentTimeProviderMock()
	blobWantListRepository := bolt.NewBlobWantListRepository(tx, currentTimeProviderMock)
	feedWantListRepository := bolt.NewFeedWantListRepository(tx, currentTimeProviderMock)
	txTestAdapters := TxTestAdapters{
		MessageRepository:     messageRepository,
		FeedRepository:        feedRepository,
		BlobRepository:        blobRepository,
		SocialGraphRepository: socialGraphRepository,
		PubRepository:         pubRepository,
		ReceiveLog:            receiveLogRepository,
		BlobWantList:          blobWantListRepository,
		FeedWantList:          feedWantListRepository,
		BanList:               banListRepository,
		CurrentTimeProvider:   currentTimeProviderMock,
		BanListHasher:         banListHasherMock,
	}
	return txTestAdapters, nil
}

var (
	_wireHopsValue = hops
)

func BuildTestAdapters(db *bbolt.DB) (TestAdapters, error) {
	private, err := identity.NewPrivate()
	if err != nil {
		return TestAdapters{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	logger := fixtures.SomeLogger()
	messageHMAC := formats.NewDefaultMessageHMAC()
	txRepositoriesFactory := newTxRepositoriesFactory(public, logger, messageHMAC)
	readMessageRepository := bolt.NewReadMessageRepository(db, txRepositoriesFactory)
	readFeedRepository := bolt.NewReadFeedRepository(db, txRepositoriesFactory)
	readReceiveLogRepository := bolt.NewReadReceiveLogRepository(db, txRepositoriesFactory)
	testAdapters := TestAdapters{
		MessageRepository: readMessageRepository,
		FeedRepository:    readFeedRepository,
		ReceiveLog:        readReceiveLogRepository,
	}
	return testAdapters, nil
}

func BuildTestCommands(t *testing.T) (TestCommands, error) {
	dialerMock := mocks.NewDialerMock()
	private, err := identity.NewPrivate()
	if err != nil {
		return TestCommands{}, err
	}
	roomsAliasRegisterHandler := commands.NewRoomsAliasRegisterHandler(dialerMock, private)
	roomsAliasRevokeHandler := commands.NewRoomsAliasRevokeHandler(dialerMock)
	peerManagerMock := mocks2.NewPeerManagerMock()
	processRoomAttendantEventHandler := commands.NewProcessRoomAttendantEventHandler(peerManagerMock)
	disconnectAllHandler := commands.NewDisconnectAllHandler(peerManagerMock)
	feedWantListRepositoryMock := mocks.NewFeedWantListRepositoryMock()
	adapters := commands.Adapters{
		FeedWantList: feedWantListRepositoryMock,
	}
	mockTransactionProvider := mocks.NewMockTransactionProvider(adapters)
	currentTimeProviderMock := mocks.NewCurrentTimeProviderMock()
	downloadFeedHandler := commands.NewDownloadFeedHandler(mockTransactionProvider, currentTimeProviderMock)
	inviteRedeemerMock := mocks.NewInviteRedeemerMock()
	logger := fixtures.TestLogger(t)
	redeemInviteHandler := commands.NewRedeemInviteHandler(inviteRedeemerMock, private, logger)
	public := privateIdentityToPublicIdentity(private)
	peerInitializerMock := mocks.NewPeerInitializerMock()
	newPeerHandlerMock := mocks.NewNewPeerHandlerMock()
	acceptTunnelConnectHandler := commands.NewAcceptTunnelConnectHandler(public, peerInitializerMock, newPeerHandlerMock)
	testCommands := TestCommands{
		RoomsAliasRegister:        roomsAliasRegisterHandler,
		RoomsAliasRevoke:          roomsAliasRevokeHandler,
		ProcessRoomAttendantEvent: processRoomAttendantEventHandler,
		DisconnectAll:             disconnectAllHandler,
		DownloadFeed:              downloadFeedHandler,
		RedeemInvite:              redeemInviteHandler,
		AcceptTunnelConnect:       acceptTunnelConnectHandler,
		PeerManager:               peerManagerMock,
		Dialer:                    dialerMock,
		FeedWantListRepository:    feedWantListRepositoryMock,
		CurrentTimeProvider:       currentTimeProviderMock,
		InviteRedeemer:            inviteRedeemerMock,
		Local:                     public,
		PeerInitializer:           peerInitializerMock,
		NewPeerHandler:            newPeerHandlerMock,
	}
	return testCommands, nil
}

func BuildTestQueries(t *testing.T) (TestQueries, error) {
	feedRepositoryMock := mocks.NewFeedRepositoryMock()
	messagePubSub := pubsub.NewMessagePubSub()
	messagePubSubMock := mocks.NewMessagePubSubMock(messagePubSub)
	logger := fixtures.TestLogger(t)
	createHistoryStreamHandler := queries.NewCreateHistoryStreamHandler(feedRepositoryMock, messagePubSubMock, logger)
	receiveLogRepositoryMock := mocks.NewReceiveLogRepositoryMock()
	receiveLogHandler := queries.NewReceiveLogHandler(receiveLogRepositoryMock)
	private, err := identity.NewPrivate()
	if err != nil {
		return TestQueries{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	publishedLogHandler, err := queries.NewPublishedLogHandler(feedRepositoryMock, receiveLogRepositoryMock, public)
	if err != nil {
		return TestQueries{}, err
	}
	messageRepositoryMock := mocks.NewMessageRepositoryMock()
	peerManagerMock := mocks2.NewPeerManagerMock()
	statusHandler := queries.NewStatusHandler(messageRepositoryMock, feedRepositoryMock, peerManagerMock)
	blobStorageMock := mocks.NewBlobStorageMock()
	getBlobHandler, err := queries.NewGetBlobHandler(blobStorageMock)
	if err != nil {
		return TestQueries{}, err
	}
	blobDownloadedPubSubMock := mocks.NewBlobDownloadedPubSubMock()
	blobDownloadedEventsHandler := queries.NewBlobDownloadedEventsHandler(blobDownloadedPubSubMock)
	dialerMock := mocks.NewDialerMock()
	roomsListAliasesHandler, err := queries.NewRoomsListAliasesHandler(dialerMock, public)
	if err != nil {
		return TestQueries{}, err
	}
	getMessageBySequenceHandler := queries.NewGetMessageBySequenceHandler(feedRepositoryMock)
	appQueries := app.Queries{
		CreateHistoryStream:  createHistoryStreamHandler,
		ReceiveLog:           receiveLogHandler,
		PublishedLog:         publishedLogHandler,
		Status:               statusHandler,
		GetBlob:              getBlobHandler,
		BlobDownloadedEvents: blobDownloadedEventsHandler,
		RoomsListAliases:     roomsListAliasesHandler,
		GetMessageBySequence: getMessageBySequenceHandler,
	}
	testQueries := TestQueries{
		Queries:              appQueries,
		FeedRepository:       feedRepositoryMock,
		MessagePubSub:        messagePubSubMock,
		MessageRepository:    messageRepositoryMock,
		PeerManager:          peerManagerMock,
		BlobStorage:          blobStorageMock,
		ReceiveLogRepository: receiveLogRepositoryMock,
		Dialer:               dialerMock,
		LocalIdentity:        public,
	}
	return testQueries, nil
}

func BuildTransactableAdapters(tx *bbolt.Tx, public identity.Public, config Config) (commands.Adapters, error) {
	graphHops := _wireGraphHopsValue
	banListHasher := adapters.NewBanListHasher()
	banListRepository := bolt.NewBanListRepository(tx, banListHasher)
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops, banListRepository)
	messageContentMappings := transport.DefaultMappings()
	logger := extractLoggerFromConfig(config)
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return commands.Adapters{}, err
	}
	messageHMAC := extractMessageHMACFromConfig(config)
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, banListRepository, scuttlebutt)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	blobWantListRepository := bolt.NewBlobWantListRepository(tx, currentTimeProvider)
	feedWantListRepository := bolt.NewFeedWantListRepository(tx, currentTimeProvider)
	commandsAdapters := commands.Adapters{
		Feed:         feedRepository,
		ReceiveLog:   receiveLogRepository,
		SocialGraph:  socialGraphRepository,
		BlobWantList: blobWantListRepository,
		FeedWantList: feedWantListRepository,
		BanList:      banListRepository,
	}
	return commandsAdapters, nil
}

var (
	_wireGraphHopsValue = hops
)

func BuildTxRepositories(tx *bbolt.Tx, public identity.Public, logger logging.Logger, messageHMAC formats.MessageHMAC) (bolt.TxRepositories, error) {
	graphHops := _wireHopsValue2
	banListHasher := adapters.NewBanListHasher()
	banListRepository := bolt.NewBanListRepository(tx, banListHasher)
	socialGraphRepository := bolt.NewSocialGraphRepository(tx, public, graphHops, banListRepository)
	messageContentMappings := transport.DefaultMappings()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return bolt.TxRepositories{}, err
	}
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageRepository := bolt.NewMessageRepository(tx, rawMessageIdentifier)
	receiveLogRepository := bolt.NewReceiveLogRepository(tx, messageRepository)
	pubRepository := bolt.NewPubRepository(tx)
	blobRepository := bolt.NewBlobRepository(tx)
	feedRepository := bolt.NewFeedRepository(tx, socialGraphRepository, receiveLogRepository, messageRepository, pubRepository, blobRepository, banListRepository, scuttlebutt)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	blobWantListRepository := bolt.NewBlobWantListRepository(tx, currentTimeProvider)
	feedWantListRepository := bolt.NewFeedWantListRepository(tx, currentTimeProvider)
	wantedFeedsRepository := bolt.NewWantedFeedsRepository(socialGraphRepository, feedWantListRepository, feedRepository, banListRepository)
	txRepositories := bolt.TxRepositories{
		Feed:         feedRepository,
		Graph:        socialGraphRepository,
		ReceiveLog:   receiveLogRepository,
		Message:      messageRepository,
		Blob:         blobRepository,
		BlobWantList: blobWantListRepository,
		FeedWantList: feedWantListRepository,
		WantedFeeds:  wantedFeedsRepository,
	}
	return txRepositories, nil
}

var (
	_wireHopsValue2 = hops
)

// BuildService creates a new service which uses the provided context as a long-term context used as a base context for
// e.g. established connections.
func BuildService(contextContext context.Context, private identity.Private, config Config) (Service, error) {
	networkKey := extractNetworkKeyFromConfig(config)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	handshaker, err := boxstream.NewHandshaker(private, networkKey, currentTimeProvider)
	if err != nil {
		return Service{}, err
	}
	requestPubSub := pubsub.NewRequestPubSub()
	connectionIdGenerator := rpc.NewConnectionIdGenerator()
	logger := extractLoggerFromConfig(config)
	peerInitializer := transport2.NewPeerInitializer(handshaker, requestPubSub, connectionIdGenerator, logger)
	dialer, err := network.NewDialer(peerInitializer, logger)
	if err != nil {
		return Service{}, err
	}
	inviteDialer := invites.NewInviteDialer(dialer, networkKey, requestPubSub, connectionIdGenerator, currentTimeProvider, logger)
	inviteRedeemer := invites2.NewInviteRedeemer(inviteDialer, logger)
	redeemInviteHandler := commands.NewRedeemInviteHandler(inviteRedeemer, private, logger)
	db, err := newBolt(config)
	if err != nil {
		return Service{}, err
	}
	public := privateIdentityToPublicIdentity(private)
	adaptersFactory := newAdaptersFactory(config, public)
	transactionProvider := bolt.NewTransactionProvider(db, adaptersFactory)
	messageContentMappings := transport.DefaultMappings()
	marshaler, err := transport.NewMarshaler(messageContentMappings, logger)
	if err != nil {
		return Service{}, err
	}
	followHandler := commands.NewFollowHandler(transactionProvider, private, marshaler, logger)
	publishRawHandler := commands.NewPublishRawHandler(transactionProvider, private, logger)
	downloadFeedHandler := commands.NewDownloadFeedHandler(transactionProvider, currentTimeProvider)
	peerManagerConfig := extractPeerManagerConfigFromConfig(config)
	tunnelDialer := tunnel.NewDialer(peerInitializer)
	sessionTracker := ebt.NewSessionTracker()
	messageHMAC := extractMessageHMACFromConfig(config)
	scuttlebutt := formats.NewScuttlebutt(marshaler, messageHMAC)
	v := newFormats(scuttlebutt)
	rawMessageIdentifier := formats.NewRawMessageIdentifier(v)
	messageBuffer := commands.NewMessageBuffer(transactionProvider, logger)
	rawMessageHandler := commands.NewRawMessageHandler(rawMessageIdentifier, messageBuffer, logger)
	txRepositoriesFactory := newTxRepositoriesFactory(public, logger, messageHMAC)
	readWantedFeedsRepository := bolt.NewReadWantedFeedsRepository(db, txRepositoriesFactory)
	wantedFeedsCache := replication.NewWantedFeedsCache(readWantedFeedsRepository)
	readFeedRepository := bolt.NewReadFeedRepository(db, txRepositoriesFactory)
	messagePubSub := pubsub.NewMessagePubSub()
	createHistoryStreamHandler := queries.NewCreateHistoryStreamHandler(readFeedRepository, messagePubSub, logger)
	createHistoryStreamHandlerAdapter := ebt2.NewCreateHistoryStreamHandlerAdapter(createHistoryStreamHandler)
	sessionRunner := ebt.NewSessionRunner(logger, rawMessageHandler, wantedFeedsCache, createHistoryStreamHandlerAdapter)
	replicator := ebt.NewReplicator(sessionTracker, sessionRunner, logger)
	manager := gossip.NewManager(logger, wantedFeedsCache)
	gossipReplicator, err := gossip.NewGossipReplicator(manager, rawMessageHandler, logger)
	if err != nil {
		return Service{}, err
	}
	negotiator := replication.NewNegotiator(logger, replicator, gossipReplicator)
	readBlobWantListRepository := bolt.NewReadBlobWantListRepository(db, txRepositoriesFactory)
	filesystemStorage, err := newFilesystemStorage(logger, config)
	if err != nil {
		return Service{}, err
	}
	blobsGetDownloader := replication2.NewBlobsGetDownloader(filesystemStorage, logger)
	blobDownloadedPubSub := pubsub.NewBlobDownloadedPubSub()
	hasHandler := replication2.NewHasHandler(filesystemStorage, readBlobWantListRepository, blobsGetDownloader, blobDownloadedPubSub, logger)
	replicationManager := replication2.NewManager(readBlobWantListRepository, filesystemStorage, hasHandler, logger)
	replicationReplicator := replication2.NewReplicator(replicationManager)
	peerRPCAdapter := rooms.NewPeerRPCAdapter(logger)
	roomAttendantEventPubSub := pubsub.NewRoomAttendantEventPubSub()
	scanner := rooms.NewScanner(peerRPCAdapter, peerRPCAdapter, roomAttendantEventPubSub, logger)
	peerManager := domain.NewPeerManager(contextContext, peerManagerConfig, dialer, tunnelDialer, negotiator, replicationReplicator, scanner, logger)
	connectHandler := commands.NewConnectHandler(peerManager, logger)
	disconnectAllHandler := commands.NewDisconnectAllHandler(peerManager)
	downloadBlobHandler := commands.NewDownloadBlobHandler(transactionProvider, currentTimeProvider)
	createBlobHandler := commands.NewCreateBlobHandler(filesystemStorage)
	addToBanListHandler := commands.NewAddToBanListHandler(transactionProvider)
	removeFromBanListHandler := commands.NewRemoveFromBanListHandler(transactionProvider)
	roomsAliasRegisterHandler := commands.NewRoomsAliasRegisterHandler(dialer, private)
	roomsAliasRevokeHandler := commands.NewRoomsAliasRevokeHandler(dialer)
	boltProgressStorage := migrations.NewBoltProgressStorage()
	runner := migrations2.NewRunner(boltProgressStorage, logger)
	goSSBRepoReader := migrations.NewGoSSBRepoReader(networkKey, messageHMAC, logger)
	migrationHandlerImportDataFromGoSSB := commands.NewMigrationHandlerImportDataFromGoSSB(goSSBRepoReader, transactionProvider, marshaler, logger)
	commandsMigrations := commands.Migrations{
		MigrationImportDataFromGoSSB: migrationHandlerImportDataFromGoSSB,
	}
	v2 := newMigrationsList(commandsMigrations, config)
	migrationsMigrations, err := migrations2.NewMigrations(v2)
	if err != nil {
		return Service{}, err
	}
	runMigrationsHandler := commands.NewRunMigrationsHandler(runner, migrationsMigrations)
	appCommands := app.Commands{
		RedeemInvite:       redeemInviteHandler,
		Follow:             followHandler,
		PublishRaw:         publishRawHandler,
		DownloadFeed:       downloadFeedHandler,
		Connect:            connectHandler,
		DisconnectAll:      disconnectAllHandler,
		DownloadBlob:       downloadBlobHandler,
		CreateBlob:         createBlobHandler,
		AddToBanList:       addToBanListHandler,
		RemoveFromBanList:  removeFromBanListHandler,
		RoomsAliasRegister: roomsAliasRegisterHandler,
		RoomsAliasRevoke:   roomsAliasRevokeHandler,
		RunMigrations:      runMigrationsHandler,
	}
	readReceiveLogRepository := bolt.NewReadReceiveLogRepository(db, txRepositoriesFactory)
	receiveLogHandler := queries.NewReceiveLogHandler(readReceiveLogRepository)
	publishedLogHandler, err := queries.NewPublishedLogHandler(readFeedRepository, readReceiveLogRepository, public)
	if err != nil {
		return Service{}, err
	}
	readMessageRepository := bolt.NewReadMessageRepository(db, txRepositoriesFactory)
	statusHandler := queries.NewStatusHandler(readMessageRepository, readFeedRepository, peerManager)
	getBlobHandler, err := queries.NewGetBlobHandler(filesystemStorage)
	if err != nil {
		return Service{}, err
	}
	blobDownloadedEventsHandler := queries.NewBlobDownloadedEventsHandler(blobDownloadedPubSub)
	roomsListAliasesHandler, err := queries.NewRoomsListAliasesHandler(dialer, public)
	if err != nil {
		return Service{}, err
	}
	getMessageBySequenceHandler := queries.NewGetMessageBySequenceHandler(readFeedRepository)
	appQueries := app.Queries{
		CreateHistoryStream:  createHistoryStreamHandler,
		ReceiveLog:           receiveLogHandler,
		PublishedLog:         publishedLogHandler,
		Status:               statusHandler,
		GetBlob:              getBlobHandler,
		BlobDownloadedEvents: blobDownloadedEventsHandler,
		RoomsListAliases:     roomsListAliasesHandler,
		GetMessageBySequence: getMessageBySequenceHandler,
	}
	application := app.Application{
		Commands: appCommands,
		Queries:  appQueries,
	}
	acceptNewPeerHandler := commands.NewAcceptNewPeerHandler(peerManager)
	listener, err := newListener(contextContext, peerInitializer, acceptNewPeerHandler, config, logger)
	if err != nil {
		return Service{}, err
	}
	discoverer, err := local.NewDiscoverer(public, logger)
	if err != nil {
		return Service{}, err
	}
	processNewLocalDiscoveryHandler := commands.NewProcessNewLocalDiscoveryHandler(peerManager)
	networkDiscoverer := network2.NewDiscoverer(discoverer, processNewLocalDiscoveryHandler, logger)
	establishNewConnectionsHandler := commands.NewEstablishNewConnectionsHandler(peerManager)
	connectionEstablisher := network2.NewConnectionEstablisher(establishNewConnectionsHandler, logger)
	handlerBlobsGet := rpc2.NewHandlerBlobsGet(getBlobHandler)
	createWantsHandler := commands.NewCreateWantsHandler(replicationManager)
	handlerBlobsCreateWants := rpc2.NewHandlerBlobsCreateWants(createWantsHandler)
	handleIncomingEbtReplicateHandler := commands.NewHandleIncomingEbtReplicateHandler(replicator)
	handlerEbtReplicate := rpc2.NewHandlerEbtReplicate(handleIncomingEbtReplicateHandler)
	acceptTunnelConnectHandler := commands.NewAcceptTunnelConnectHandler(public, peerInitializer, peerManager)
	handlerTunnelConnect := rpc2.NewHandlerTunnelConnect(acceptTunnelConnectHandler)
	v3 := rpc2.NewMuxHandlers(handlerBlobsGet, handlerBlobsCreateWants, handlerEbtReplicate, handlerTunnelConnect)
	handlerCreateHistoryStream := rpc2.NewHandlerCreateHistoryStream(createHistoryStreamHandler, logger)
	v4 := rpc2.NewMuxClosingHandlers(handlerCreateHistoryStream)
	muxMux, err := mux.NewMux(logger, v3, v4)
	if err != nil {
		return Service{}, err
	}
	requestSubscriber := pubsub2.NewRequestSubscriber(requestPubSub, muxMux)
	processRoomAttendantEventHandler := commands.NewProcessRoomAttendantEventHandler(peerManager)
	roomAttendantEventSubscriber := pubsub2.NewRoomAttendantEventSubscriber(roomAttendantEventPubSub, processRoomAttendantEventHandler, logger)
	advertiser, err := newAdvertiser(public, config)
	if err != nil {
		return Service{}, err
	}
	service := NewService(application, listener, networkDiscoverer, connectionEstablisher, requestSubscriber, roomAttendantEventSubscriber, advertiser, messageBuffer, createHistoryStreamHandler)
	return service, nil
}

// wire.go:

type TxTestAdapters struct {
	MessageRepository     *bolt.MessageRepository
	FeedRepository        *bolt.FeedRepository
	BlobRepository        *bolt.BlobRepository
	SocialGraphRepository *bolt.SocialGraphRepository
	PubRepository         *bolt.PubRepository
	ReceiveLog            *bolt.ReceiveLogRepository
	BlobWantList          *bolt.BlobWantListRepository
	FeedWantList          *bolt.FeedWantListRepository
	BanList               *bolt.BanListRepository

	CurrentTimeProvider *mocks.CurrentTimeProviderMock
	BanListHasher       *mocks.BanListHasherMock
}

type TestAdapters struct {
	MessageRepository *bolt.ReadMessageRepository
	FeedRepository    *bolt.ReadFeedRepository
	ReceiveLog        *bolt.ReadReceiveLogRepository
}

type TestCommands struct {
	RoomsAliasRegister        *commands.RoomsAliasRegisterHandler
	RoomsAliasRevoke          *commands.RoomsAliasRevokeHandler
	ProcessRoomAttendantEvent *commands.ProcessRoomAttendantEventHandler
	DisconnectAll             *commands.DisconnectAllHandler
	DownloadFeed              *commands.DownloadFeedHandler
	RedeemInvite              *commands.RedeemInviteHandler
	AcceptTunnelConnect       *commands.AcceptTunnelConnectHandler

	PeerManager            *mocks2.PeerManagerMock
	Dialer                 *mocks.DialerMock
	FeedWantListRepository *mocks.FeedWantListRepositoryMock
	CurrentTimeProvider    *mocks.CurrentTimeProviderMock
	InviteRedeemer         *mocks.InviteRedeemerMock
	Local                  identity.Public
	PeerInitializer        *mocks.PeerInitializerMock
	NewPeerHandler         *mocks.NewPeerHandlerMock
}

type TestQueries struct {
	Queries app.Queries

	FeedRepository       *mocks.FeedRepositoryMock
	MessagePubSub        *mocks.MessagePubSubMock
	MessageRepository    *mocks.MessageRepositoryMock
	PeerManager          *mocks2.PeerManagerMock
	BlobStorage          *mocks.BlobStorageMock
	ReceiveLogRepository *mocks.ReceiveLogRepositoryMock
	Dialer               *mocks.DialerMock

	LocalIdentity identity.Public
}

var replicatorSet = wire.NewSet(gossip.NewManager, wire.Bind(new(gossip.ReplicationManager), new(*gossip.Manager)), gossip.NewGossipReplicator, wire.Bind(new(replication.CreateHistoryStreamReplicator), new(*gossip.GossipReplicator)), ebt.NewReplicator, wire.Bind(new(replication.EpidemicBroadcastTreesReplicator), new(ebt.Replicator)), replication.NewWantedFeedsCache, wire.Bind(new(gossip.ContactsStorage), new(*replication.WantedFeedsCache)), wire.Bind(new(ebt.ContactsStorage), new(*replication.WantedFeedsCache)), ebt.NewSessionTracker, wire.Bind(new(ebt.Tracker), new(*ebt.SessionTracker)), ebt.NewSessionRunner, wire.Bind(new(ebt.Runner), new(*ebt.SessionRunner)), replication.NewNegotiator, wire.Bind(new(domain.MessageReplicator), new(*replication.Negotiator)))

var blobReplicatorSet = wire.NewSet(replication2.NewManager, wire.Bind(new(replication2.ReplicationManager), new(*replication2.Manager)), wire.Bind(new(commands.BlobReplicationManager), new(*replication2.Manager)), replication2.NewReplicator, wire.Bind(new(domain.BlobReplicator), new(*replication2.Replicator)), replication2.NewBlobsGetDownloader, wire.Bind(new(replication2.Downloader), new(*replication2.BlobsGetDownloader)), replication2.NewHasHandler, wire.Bind(new(replication2.HasBlobHandler), new(*replication2.HasHandler)))

var hops = graph.MustNewHops(3)

func newAdvertiser(l identity.Public, config Config) (*local.Advertiser, error) {
	return local.NewAdvertiser(l, config.ListenAddress)
}

func newAdaptersFactory(config Config, local2 identity.Public) bolt.AdaptersFactory {
	return func(tx *bbolt.Tx) (commands.Adapters, error) {
		return BuildTransactableAdapters(tx, local2, config)
	}
}

func newBolt(config Config) (*bbolt.DB, error) {
	filename := path.Join(config.DataDirectory, "database.bolt")
	b, err := bbolt.Open(filename, 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database, is something else reading it?")
	}
	return b, nil
}

func privateIdentityToPublicIdentity(p identity.Private) identity.Public {
	return p.Public()
}
