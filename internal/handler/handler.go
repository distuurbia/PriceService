// Package handler contains handler methods of grpc to handle requests
package handler

import (
	"context"

	"github.com/distuurbia/PriceService/internal/model"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PriceServiceService is an interface that contains methods of PriceService service.
type PriceServiceService interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
	AddSubscriber(subscriberID uuid.UUID, selectedShares []string) error
	DeleteSubscriber(subscriberID uuid.UUID) error
	SendToSubscriber(ctx context.Context, subscriberID uuid.UUID) ([]*protocol.Share, error)
	SendToAllSubscribedChans(ctx context.Context)
}

// Handler is responsible for handling gRPC requests related to entities.
type Handler struct {
	s PriceServiceService
	protocol.UnimplementedPriceServiceServiceServer
}

// NewHandler creates a new instance of the Handler struct.
func NewHandler(s PriceServiceService) *Handler {
	return &Handler{
		s: s,
	}
}

// Subscribe takes message from redis stream through PriceServiceService and sends it to grpc stream.
func (h *Handler) Subscribe(req *protocol.SubscribeRequest, stream protocol.PriceServiceService_SubscribeServer) error {
	subscriberID, err := uuid.Parse(req.UUID)
	if err != nil {
		logrus.Errorf("Handler -> ReadFromStream -> uuid.Parse: %v", err)
		return err
	}

	err = h.s.AddSubscriber(subscriberID, req.SelectedShares)
	if err != nil {
		logrus.Errorf("Handler -> ReadFromStream -> AddSubscriber: %v", err)
		return err
	}

	for {
		protoShares, errSend := h.s.SendToSubscriber(stream.Context(), subscriberID)

		if errSend != nil {
			logrus.Errorf("Handler -> Subscribe -> SendToSubscriber -> %v", err)

			errDelete := h.s.DeleteSubscriber(subscriberID)
			if errDelete != nil {
				logrus.Errorf("Handler -> DeleteSubscriber: %v", err)
				return errDelete
			}

			return errSend
		}
		err := stream.Send(&protocol.SubscribeResponse{Shares: protoShares})
		if err != nil {
			logrus.Errorf("Handler -> Subscribe -> stream.Send: %v", err)
			return err
		}
	}
}
