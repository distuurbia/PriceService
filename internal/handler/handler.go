// Package handler contains handler methods of grpc to handle requests
package handler

import (
	"context"

	"github.com/distuurbia/PriceService/internal/model"
	"github.com/distuurbia/PriceService/proto_services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PriceServiceService is an interface that contains methods of PriceService service.
type PriceServiceService interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
	AddSubscriber(subscriberID uuid.UUID, selectedShares []string)
	DeleteSubscriber(subscriberID uuid.UUID)
	SendToSubscriber(ctx context.Context, subscriberID uuid.UUID, stream proto_services.PriceServiceService_SubscribeServer)
	SendToAllSubscribedChans(ctx context.Context)
}

// Handler is responsible for handling gRPC requests related to entities.
type Handler struct {
	priceServiceSrv PriceServiceService
	proto_services.UnimplementedPriceServiceServiceServer
}

// NewHandler creates a new instance of the Handler struct.
func NewHandler(priceServiceSrv PriceServiceService) *Handler {
	return &Handler{
		priceServiceSrv: priceServiceSrv,
	}
}

// Subscribe takes message from redis stream through PriceServiceService and sends it to grpc stream.
func (handl *Handler) Subscribe(req *proto_services.SubscribeRequest, stream proto_services.PriceServiceService_SubscribeServer) error {
	subscriberID, err := uuid.Parse(req.UUID)
	if err != nil {
		logrus.Errorf("Handler -> ReadFromStream -> uuid.Parse: %v", err)
		return err
	}

	handl.priceServiceSrv.AddSubscriber(subscriberID, req.SelectedShares)
	handl.priceServiceSrv.SendToSubscriber(stream.Context(), subscriberID, stream)
	handl.priceServiceSrv.DeleteSubscriber(subscriberID)

	return nil
}
