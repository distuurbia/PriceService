package service

import (
	"context"
	"testing"

	"github.com/distuurbia/PriceService/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddSubscriber(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	s := NewPriceServiceService(r)

	err := s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	selectedShares, ok := s.submngr.Subscribers[testSubID]
	require.True(t, ok)
	require.Equal(t, len(selectedShares), len(testSelectedShares))
}

func TestDeleteSubscriber(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	s := NewPriceServiceService(r)

	err := s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	err = s.DeleteSubscriber(testSubID)
	require.NoError(t, err)

	err = s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)
}

func TestReadFromStream(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	r.On("ReadFromStream", mock.Anything).
		Return(testShares, nil).
		Once()
	s := NewPriceServiceService(r)
	shares, err := s.ReadFromStream(context.Background())
	require.NoError(t, err)
	require.Equal(t, len(shares), len(testShares))
	r.AssertExpectations(t)
}

func TestSendToAllSubscribedChans(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	r.On("ReadFromStream", mock.Anything).
		Return(testShares, nil)

	s := NewPriceServiceService(r)

	err := s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go s.SendToAllSubscribedChans(ctx)

	readedShares := make(map[string]float64)
	for i := 0; i < 3; i++ {
		testShare := <-s.submngr.SubscribersShares[testSubID]
		readedShares[testShare.Name] = testShare.Price
	}
	cancel()

	err = s.DeleteSubscriber(testSubID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.Equal(t, testShares[i].Price, readedShares[testShares[i].Name])
	}
	r.AssertExpectations(t)
}

func TestSendToSubscriber(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	r.On("ReadFromStream", mock.Anything).
		Return(testShares, nil)

	s := NewPriceServiceService(r)

	ctx, cancel := context.WithCancel(context.Background())
	go s.SendToAllSubscribedChans(ctx)

	err := s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	shares, err := s.SendToSubscriber(ctx, testSubID)
	require.NoError(t, err)

	cancel()

	err = s.DeleteSubscriber(testSubID)
	require.NoError(t, err)

	require.Equal(t, len(testShares), len(shares))

	r.AssertExpectations(t)
}
