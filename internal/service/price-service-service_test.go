package service

import (
	"context"
	"testing"
	"time"

	"github.com/distuurbia/PriceService/internal/service/mocks"
	"github.com/distuurbia/PriceService/proto_services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	selectedShares, ok := priceServiceSrv.subscribersMngr.Subscribers[testSubID]
	require.True(t, ok)
	require.Equal(t, len(selectedShares), len(testSelectedShares))

}

func TestAddNilIDSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(uuid.Nil, testSelectedShares)
	require.Error(t, err)
}

func TestAddEmptySharesSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, []string{})
	require.Error(t, err)
}

func TestAddExistingSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	err = priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.Error(t, err)
}

func TestDeleteSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	err = priceServiceSrv.DeleteSubscriber(testSubID)
	require.NoError(t, err)
}

func TestDeleteNilIDSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.DeleteSubscriber(uuid.Nil)
	require.Error(t, err)
}

func TestDeleteNotExistingSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.DeleteSubscriber(testSubID)
	require.Error(t, err)
}

func TestReadFromStream(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceRepo.On("ReadFromStream", mock.Anything).
		Return(testShares, nil).
		Once()
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)
	shares, err := priceServiceSrv.ReadFromStream(context.Background())
	require.NoError(t, err)
	require.Equal(t, len(shares), len(testShares))
	priceServiceRepo.AssertExpectations(t)
}

func TestSendToAllSubscribedChans(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceRepo.On("ReadFromStream", mock.Anything).
		Return(testShares, nil)

	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	priceServiceSrv.SendToAllSubscribedChans(ctx)

	close(priceServiceSrv.subscribersMngr.SubscribersShares[testSubID])
	shares := <-priceServiceSrv.subscribersMngr.SubscribersShares[testSubID]
	cancel()
	require.Equal(t, len(testSelectedShares), len(shares))
	
	priceServiceRepo.AssertExpectations(t)
}

func TestSendToSubscriber(t *testing.T) {
	priceServiceRepo := new(mocks.PriceServiceRepository)
	priceServiceRepo.On("SendToSubscriber", mock.Anything, mock.Anything).
		Return(nil)
	priceServiceSrv := NewPriceServiceService(priceServiceRepo)

	err := priceServiceSrv.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	priceServiceSrv.subscribersMngr.SubscribersShares[testSubID] <- testShares

	close(priceServiceSrv.subscribersMngr.SubscribersShares[testSubID])

	var stream proto_services.PriceServiceService_SubscribeServer

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = priceServiceSrv.SendToSubscriber(ctx, testSubID, stream)
	cancel()
	require.NoError(t, err)
	
	priceServiceRepo.AssertExpectations(t)

	

}