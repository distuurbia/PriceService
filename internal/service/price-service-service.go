// Package service contains bisnes logic of Price Service
package service

import (
	"context"
	"fmt"

	"github.com/distuurbia/PriceService/internal/model"
)

// PriceServiceRepository is an interface of PriceServiceRepository structure of repository
type PriceServiceRepository interface {
	ReadFromStream(ctx context.Context) (shares []*model.Share, err error)
}

// PriceServiceService contains an inerface of PriceServiceRepository
type PriceServiceService struct {
	priceServiceRepo PriceServiceRepository
}

// NewPriceServiceService creates an object of PriceServiceService by using PriceServiceRepository interface
func NewPriceServiceService(priceServiceRepo PriceServiceRepository) *PriceServiceService {
	return &PriceServiceService{priceServiceRepo: priceServiceRepo}
}

// ReadFromStream reads last message from the redis stream using repository ReadFromStream method
func (priceServiceSrv *PriceServiceService) ReadFromStream(ctx context.Context) (shares []*model.Share, err error) {
	shares, err = priceServiceSrv.priceServiceRepo.ReadFromStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("PriceServiceService -> ReadFromStream -> %w", err)
	}
	return shares, nil
}
