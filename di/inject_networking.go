package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/rooms/tunnel"
	domaintransport "github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	portsnetwork "github.com/planetary-social/scuttlego/service/ports/network"
)

//nolint:unused
var networkingSet = wire.NewSet(
	domaintransport.NewPeerInitializer,
	wire.Bind(new(portsnetwork.ServerPeerInitializer), new(*domaintransport.PeerInitializer)),
	wire.Bind(new(network.ClientPeerInitializer), new(*domaintransport.PeerInitializer)),
	wire.Bind(new(tunnel.ClientPeerInitializer), new(*domaintransport.PeerInitializer)),

	rpc.NewConnectionIdGenerator,

	boxstream.NewHandshaker,

	network.NewDialer,
	wire.Bind(new(commands.Dialer), new(*network.Dialer)),
	wire.Bind(new(queries.Dialer), new(*network.Dialer)),
	wire.Bind(new(domain.Dialer), new(*network.Dialer)),
)
