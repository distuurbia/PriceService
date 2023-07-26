// Package handler contains handler methods of grpc to handle requests
package handler

import (
	"context"

	"github.com/distuurbia/PriceService/internal/model"
	"github.com/distuurbia/PriceService/proto_services"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
)

// PriceServiceService is an interface that contains methods of PriceService service.
type PriceServiceService interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
}

// GRPCHandler is responsible for handling gRPC requests related to entities.
type GRPCHandler struct {
	priceServiceSrv PriceServiceService
	validate        *validator.Validate
	proto_services.UnimplementedPriceServiceServiceServer
}

// NewGRPCHandler creates a new instance of the GRPCHandler struct.
func NewGRPCHandler(priceServiceSrv PriceServiceService, v *validator.Validate) *GRPCHandler {
	return &GRPCHandler{
		priceServiceSrv: priceServiceSrv,
		validate:        v,
	}
}

// ReadFromStream takes message from redis stream through PriceServiceService and sends it to grpc stream.
func (h *GRPCHandler) ReadFromStream(_ *proto_services.ReadFromStreamRequest, stream proto_services.PriceServiceService_ReadFromStreamServer) error {
	for {
		shares, err := h.priceServiceSrv.ReadFromStream(context.Background())
		if err != nil {
			logrus.Errorf("GRPCHandler -> ReadFromStream: %v", err)
			return err
		}
		var protoShares []*proto_services.Share
		for _, share := range shares {
			protoShares = append(protoShares, &proto_services.Share{
				Name:  share.Name,
				Price: share.Price,
			})
		}
		err = stream.Send(&proto_services.ReadFromStreamResponse{Shares: protoShares})
		if err != nil {
			logrus.Errorf("GRPCHandler -> ReadFromStream -> stream.Send: %v", err)
			return err
		}
	}
}
