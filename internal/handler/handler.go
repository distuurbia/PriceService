// Package handler contains handler methods of grpc to handle requests
package handler

import (
	"context"
	"fmt"

	"github.com/distuurbia/PriceService/internal/model"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/go-playground/validator"
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
	s        PriceServiceService
	validate *validator.Validate
	protocol.UnimplementedPriceServiceServiceServer
}

// NewHandler creates a new instance of the Handler struct.
func NewHandler(s PriceServiceService, validate *validator.Validate) *Handler {
	return &Handler{
		s:        s,
		validate: validate,
	}
}

// ValidationID validate given in and parses it to uuid.UUID type
func (h *Handler) ValidationID(ctx context.Context, id string) (uuid.UUID, error) {
	err := h.validate.VarCtx(ctx, id, "required,uuid")
	if err != nil {
		logrus.Errorf("ValidationID -> %v", err)
		return uuid.Nil, err
	}

	validatedID, err := uuid.Parse(id)
	if err != nil {
		logrus.Errorf("ValidationID -> %v", err)
		return uuid.Nil, err
	}

	if validatedID == uuid.Nil {
		logrus.Errorf("ValidationID -> error: failed to use uuid")
		return uuid.Nil, fmt.Errorf("ValidationID -> error: failed to use uuid")
	}

	return validatedID, nil
}

// Subscribe takes message from redis stream through PriceServiceService and sends it to grpc stream.
func (h *Handler) Subscribe(req *protocol.SubscribeRequest, stream protocol.PriceServiceService_SubscribeServer) error {
	subscriberID, err := h.ValidationID(stream.Context(), req.UUID)
	if err != nil {
		logrus.Errorf("Handler -> Subscribe -> %v", err)
		return err
	}

	err = h.validate.VarCtx(stream.Context(), req.SelectedShares, "required")
	if err != nil {
		logrus.Errorf("Handler -> Subscribe -> %v", err)
		return err
	}

	err = h.s.AddSubscriber(subscriberID, req.SelectedShares)
	if err != nil {
		logrus.Errorf("Handler -> Subscribe -> %v", err)
		return err
	}

	for {
		protoShares, errSend := h.s.SendToSubscriber(stream.Context(), subscriberID)

		if errSend != nil {
			logrus.Errorf("Handler -> Subscribe -> %v", err)

			errDelete := h.s.DeleteSubscriber(subscriberID)
			if errDelete != nil {
				logrus.Errorf("Handler -> Subscribe -> %v", err)
				return errDelete
			}

			return errSend
		}
		err := stream.Send(&protocol.SubscribeResponse{Shares: protoShares})
		if err != nil {
			logrus.Errorf("Handler -> Subscribe -> %v", err)
			return err
		}
	}
}
