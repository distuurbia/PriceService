// Package service contains bisnes logic of Price Service
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/distuurbia/PriceService/internal/model"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PriceServiceRepository is an interface of PriceServiceRepository structure of repository
type PriceServiceRepository interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
}

// PriceServiceService contains an inerface of PriceServiceRepository
type PriceServiceService struct {
	r       PriceServiceRepository
	submngr model.SubscribersManager
}

// NewPriceServiceService creates an object of PriceServiceService by using PriceServiceRepository interface
func NewPriceServiceService(r PriceServiceRepository) *PriceServiceService {
	return &PriceServiceService{r: r,
		submngr: model.SubscribersManager{SubscribersShares: make(map[uuid.UUID]chan []*model.Share),
			Subscribers: make(map[uuid.UUID][]string)}}
}

// AddSubscriber adds new subscriber to subscribe map in SubscriberManager
func (s *PriceServiceService) AddSubscriber(subscriberID uuid.UUID, selectedShares []string) error {
	const msgs = 1
	s.submngr.Mu.Lock()
	defer s.submngr.Mu.Unlock()
	if _, ok := s.submngr.Subscribers[subscriberID]; !ok {
		s.submngr.Subscribers[subscriberID] = selectedShares
		s.submngr.SubscribersShares[subscriberID] = make(chan []*model.Share, msgs)
		return nil
	}
	return fmt.Errorf("PriceServiceService -> AddSubscriber -> error: subscriber with such ID already exists")
}

// DeleteSubscriber delete subscriber from subscribe map in SubscriberManager by uuid
func (s *PriceServiceService) DeleteSubscriber(subscriberID uuid.UUID) error {
	s.submngr.Mu.Lock()
	defer s.submngr.Mu.Unlock()
	if _, ok := s.submngr.Subscribers[subscriberID]; ok {
		delete(s.submngr.Subscribers, subscriberID)
		close(s.submngr.SubscribersShares[subscriberID])
		delete(s.submngr.SubscribersShares, subscriberID)
		return nil
	}
	return fmt.Errorf("PriceServiceService -> DeleteSubscriber -> error: subscriber with such ID doesn't exists")
}

// ReadFromStream reads last message from the redis stream using repository ReadFromStream method
func (s *PriceServiceService) ReadFromStream(ctx context.Context) (shares []*model.Share, err error) {
	shares, err = s.r.ReadFromStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("PriceServiceService -> ReadFromStream -> %w", err)
	}
	return shares, nil
}

// SendToAllSubscribedChans sends in loop actual info about subscribed shares to subscribers chans
func (s *PriceServiceService) SendToAllSubscribedChans(ctx context.Context) {
	for {
		if len(s.submngr.Subscribers) > 0 {
			shares, err := s.ReadFromStream(ctx)
			if err != nil {
				logrus.Errorf("PriceServiceService -> SendToAllSubscribedChans: %v", err)
				return
			}
			for subID, selcetedShares := range s.submngr.Subscribers {
				tempShares := make([]*model.Share, 0)
				for _, share := range shares {
					if strings.Contains(strings.Join(selcetedShares, ","), share.Name) {
						tempShares = append(tempShares, share)
					}
				}

				select {
				case <-ctx.Done():
					return
				case s.submngr.SubscribersShares[subID] <- tempShares:
				}
			}
		}
	}
}

// SendToSubscriber calls SendToSubscriber method of repository
func (s *PriceServiceService) SendToSubscriber(ctx context.Context, subscriberID uuid.UUID) (protoShares []*protocol.Share, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case shares := <-s.submngr.SubscribersShares[subscriberID]:
		for _, share := range shares {
			protoShares = append(protoShares, &protocol.Share{
				Name:  share.Name,
				Price: share.Price,
			})
		}
		return protoShares, nil
	}
}
