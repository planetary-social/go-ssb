package rpc

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type CreateWantsCommandHandler interface {
	Handle(ctx context.Context, cmd commands.CreateWants) (<-chan messages.BlobWithSizeOrWantDistance, error)
}

type HandlerBlobsCreateWants struct {
	handler CreateWantsCommandHandler
}

func NewHandlerBlobsCreateWants(handler CreateWantsCommandHandler) *HandlerBlobsCreateWants {
	return &HandlerBlobsCreateWants{
		handler: handler,
	}
}

func (h HandlerBlobsCreateWants) Procedure() rpc.Procedure {
	return messages.BlobsCreateWantsProcedure
}

func (h HandlerBlobsCreateWants) Handle(ctx context.Context, s mux.Stream, req *rpc.Request) error {
	cmd := commands.CreateWants{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := h.handler.Handle(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "could not execute the create wants command")
	}

	for v := range ch {
		resp, err := messages.NewBlobsCreateWantsResponse(v.Id(), v.SizeOrWantDistance())
		if err != nil {
			return errors.Wrap(err, "failed to create a response")
		}

		j, err := resp.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "json marshalling failed")
		}

		if err := s.WriteMessage(j, transport.MessageBodyTypeJSON); err != nil {
			return errors.Wrap(err, "failed to send a message")
		}
	}

	return nil
}
