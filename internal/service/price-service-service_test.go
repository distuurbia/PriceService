package service

import (
	"context"
	"testing"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	s.SendToAllSubscribedChans(ctx)

	close(s.submngr.SubscribersShares[testSubID])
	shares := <-s.submngr.SubscribersShares[testSubID]
	cancel()
	require.Equal(t, len(testSelectedShares), len(shares))

	r.AssertExpectations(t)
}

func TestSendToSubscriber(t *testing.T) {
	r := new(mocks.PriceServiceRepository)
	s := NewPriceServiceService(r)

	err := s.AddSubscriber(testSubID, testSelectedShares)
	require.NoError(t, err)

	s.submngr.SubscribersShares[testSubID] <- testShares

	close(s.submngr.SubscribersShares[testSubID])

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	shares, err := s.SendToSubscriber(ctx, testSubID)
	cancel()
	require.NoError(t, err)

	require.Equal(t, len(testShares), len(shares))

	r.AssertExpectations(t)
}
