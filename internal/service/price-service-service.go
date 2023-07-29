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
	SendToSubscriber(ctx context.Context, subscriberID uuid.UUID, subscribersShares map[uuid.UUID]chan []*model.Share, stream proto_services.PriceServiceService_SubscribeServer)
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
func (priceServiceSrv *PriceServiceService) AddSubscriber(subscriberID uuid.UUID, selectedShares []string) {
	priceServiceSrv.subscribersMngr.Mu.Lock()
	defer priceServiceSrv.subscribersMngr.Mu.Unlock()
	priceServiceSrv.subscribersMngr.Subscribers[subscriberID] = selectedShares
}

// DeleteSubscriber delete subscriber from subscribe map in SubscriberManager by uuid
func (priceServiceSrv *PriceServiceService) DeleteSubscriber(subscriberID uuid.UUID) {
	priceServiceSrv.subscribersMngr.Mu.Lock()
	defer priceServiceSrv.subscribersMngr.Mu.Unlock()
	delete(priceServiceSrv.subscribersMngr.Subscribers, subscriberID)
}

// GetSubscribers gets subscribers from subscribers manager
func (priceServiceSrv *PriceServiceService) GetSubscribers() map[uuid.UUID][]string {
	priceServiceSrv.subscribersMngr.Mu.Lock()
	defer priceServiceSrv.subscribersMngr.Mu.Unlock()
	return priceServiceSrv.subscribersMngr.Subscribers
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
	const msgs = 100
	for {
		if len(priceServiceSrv.subscribersMngr.Subscribers) > 0 {
			shares, err := priceServiceSrv.ReadFromStream(ctx)
			if err != nil {
				logrus.Errorf("PriceServiceService -> SendToAllSubscribedChans: %v", err)
				return
			}
			for subID, selcetedShares := range priceServiceSrv.GetSubscribers() {
				tempShares := make([]*model.Share, 0)
				for _, share := range shares {
					if strings.Contains(strings.Join(selcetedShares, ","), share.Name) {
						tempShares = append(tempShares, share)
					}
				}
				priceServiceSrv.subscribersMngr.SubscribersShares[subID] = make(chan []*model.Share, msgs)
				priceServiceSrv.subscribersMngr.SubscribersShares[subID] <- tempShares
			}
		}
	}
}

// SendToSubscriber calls SendToSubscriber method of repository
func (priceServiceSrv *PriceServiceService) SendToSubscriber(ctx context.Context, subscriberID uuid.UUID, stream proto_services.PriceServiceService_SubscribeServer) {
	priceServiceSrv.priceServiceRepo.SendToSubscriber(ctx, subscriberID, priceServiceSrv.subscribersMngr.SubscribersShares, stream)
}
