package pubsub

import (
	"context"

	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type RoomAttendantEventPubSub struct {
	pubsub *GoChannelPubSub[RoomAttendantEvent]
}

func NewRoomAttendantEventPubSub() *RoomAttendantEventPubSub {
	return &RoomAttendantEventPubSub{
		pubsub: NewGoChannelPubSub[RoomAttendantEvent](),
	}
}

func (p *RoomAttendantEventPubSub) PublishAttendantEvent(portal transport.Peer, event rooms.RoomAttendantsEvent) error {
	p.pubsub.Publish(RoomAttendantEvent{
		Portal: portal,
		Event:  event,
	})
	return nil
}

func (p *RoomAttendantEventPubSub) SubscribeToAttendantEvents(ctx context.Context) <-chan RoomAttendantEvent {
	return p.pubsub.Subscribe(ctx)
}

type RoomAttendantEvent struct {
	Portal transport.Peer
	Event  rooms.RoomAttendantsEvent
}
