// Package service contains bisnes logic of Price Service
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/distuurbia/PriceService/internal/model"
	"github.com/distuurbia/PriceService/proto_services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PriceServiceRepository is an interface of PriceServiceRepository structure of repository
type PriceServiceRepository interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
	SendToSubscriber(protoShares []*proto_services.Share,
		stream proto_services.PriceServiceService_SubscribeServer) error
}

// PriceServiceService contains an inerface of PriceServiceRepository
type PriceServiceService struct {
	priceServiceRepo PriceServiceRepository
	subscribersMngr  model.SubscribersManager
}

// NewPriceServiceService creates an object of PriceServiceService by using PriceServiceRepository interface
func NewPriceServiceService(priceServiceRepo PriceServiceRepository) *PriceServiceService {
	return &PriceServiceService{priceServiceRepo: priceServiceRepo,
		subscribersMngr: model.SubscribersManager{SubscribersShares: make(map[uuid.UUID]chan []*model.Share),
			Subscribers: make(map[uuid.UUID][]string)}}
}

// AddSubscriber adds new subscriber to subscribe map in SubscriberManager
func (priceServiceSrv *PriceServiceService) AddSubscriber(subscriberID uuid.UUID, selectedShares []string) error {
	const msgs = 1
	priceServiceSrv.subscribersMngr.Mu.Lock()
	defer priceServiceSrv.subscribersMngr.Mu.Unlock()
	if subscriberID == uuid.Nil {
		return fmt.Errorf("PriceServiceService -> AddSubscriber -> error: subscriber has nil uuid")
	}
	if len(selectedShares) == 0 {
		return fmt.Errorf("PriceServiceService -> AddSubscriber -> error: subscriber hasn't subscribed on any shares")
	}
	if _, ok := priceServiceSrv.subscribersMngr.Subscribers[subscriberID]; !ok {
		priceServiceSrv.subscribersMngr.Subscribers[subscriberID] = selectedShares
		priceServiceSrv.subscribersMngr.SubscribersShares[subscriberID] = make(chan []*model.Share, msgs)
		return nil
	}
	return fmt.Errorf("PriceServiceService -> AddSubscriber -> error: subscriber with such ID already exists")
}

// DeleteSubscriber delete subscriber from subscribe map in SubscriberManager by uuid
func (priceServiceSrv *PriceServiceService) DeleteSubscriber(subscriberID uuid.UUID) error {
	priceServiceSrv.subscribersMngr.Mu.Lock()
	defer priceServiceSrv.subscribersMngr.Mu.Unlock()
	if subscriberID == uuid.Nil {
		return fmt.Errorf("PriceServiceService -> DeleteSubscriber -> error: subscriber has nil uuid")
	}
	if _, ok := priceServiceSrv.subscribersMngr.Subscribers[subscriberID]; ok {
		delete(priceServiceSrv.subscribersMngr.Subscribers, subscriberID)
		close(priceServiceSrv.subscribersMngr.SubscribersShares[subscriberID])
		delete(priceServiceSrv.subscribersMngr.SubscribersShares, subscriberID)
		return nil
	}
	return fmt.Errorf("PriceServiceService -> DeleteSubscriber -> error: subscriber with such ID doesn't exists")
}

// ReadFromStream reads last message from the redis stream using repository ReadFromStream method
func (priceServiceSrv *PriceServiceService) ReadFromStream(ctx context.Context) (shares []*model.Share, err error) {
	shares, err = priceServiceSrv.priceServiceRepo.ReadFromStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("PriceServiceService -> ReadFromStream -> %w", err)
	}
	return shares, nil
}

// SendToAllSubscribedChans sends in loop actual info about subscribed shares to subscribers chans
func (priceServiceSrv *PriceServiceService) SendToAllSubscribedChans(ctx context.Context) {
	for {
		if len(priceServiceSrv.subscribersMngr.Subscribers) > 0 {
			shares, err := priceServiceSrv.ReadFromStream(ctx)
			if err != nil {
				logrus.Errorf("PriceServiceService -> SendToAllSubscribedChans: %v", err)
				return
			}
			for subID, selcetedShares := range priceServiceSrv.subscribersMngr.Subscribers {
				tempShares := make([]*model.Share, 0)
				for _, share := range shares {
					if strings.Contains(strings.Join(selcetedShares, ","), share.Name) {
						tempShares = append(tempShares, share)
					}
				}

				select {
					case <- ctx.Done():
						return
					case priceServiceSrv.subscribersMngr.SubscribersShares[subID] <- tempShares:
				}
				
			}
		}

	}
}

// SendToSubscriber calls SendToSubscriber method of repository
func (priceServiceSrv *PriceServiceService) SendToSubscriber(ctx context.Context, subscriberID uuid.UUID, stream proto_services.PriceServiceService_SubscribeServer) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case shares := <-priceServiceSrv.subscribersMngr.SubscribersShares[subscriberID]:
			var protoShares []*proto_services.Share
			for _, share := range shares {
				protoShares = append(protoShares, &proto_services.Share{
					Name:  share.Name,
					Price: share.Price,
				})
			}

			err := priceServiceSrv.priceServiceRepo.SendToSubscriber(protoShares, stream)
			if err != nil {
				return fmt.Errorf("PriceServiceRepository -> SendToSubscriber -> %v", err)
			}
		}
	}
}
