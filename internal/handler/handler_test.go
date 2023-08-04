package handler

import (
	"context"
	"testing"

	"github.com/distuurbia/PriceService/internal/handler/mocks"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubscribe(t *testing.T) {
	s := new(mocks.PriceServiceService)

	s.On("AddSubscriber", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("[]string")).
		Return(nil)
	s.On("DeleteSubscriber", mock.AnythingOfType("uuid.UUID")).
		Return(nil)
	s.On("SendToSubscriber", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(testShares, nil)
	client, closer := Server(context.Background(), s)

	type expectation struct {
		subResponses []*protocol.SubscribeResponse
		err          error
	}
	respProtoShares := make([]*protocol.Share, 0)
	respProtoShares = append(respProtoShares,
		&protocol.Share{Name: "Apple", Price: 250},
		&protocol.Share{Name: "Tesla", Price: 1000})

	reqSelectedShares := make([]string, 0)
	reqSelectedShares = append(reqSelectedShares, "Apple", "Tesla")

	testReqResp := struct {
		subReq       *protocol.SubscribeRequest
		expectedResp expectation
	}{subReq: &protocol.SubscribeRequest{UUID: "747b6b85-9441-48cd-aee5-932f386ba381", SelectedShares: reqSelectedShares},
		expectedResp: expectation{
			subResponses: []*protocol.SubscribeResponse{
				{Shares: respProtoShares},
				{Shares: respProtoShares},
			},
			err: nil,
		},
	}

	out, err := client.Subscribe(context.Background(), testReqResp.subReq)
	require.NoError(t, err)
	var outs []*protocol.SubscribeResponse

	for i := 0; i < 2; i++ {
		o, err := out.Recv()
		require.NoError(t, err)

		outs = append(outs, o)
	}

	require.Equal(t, len(testReqResp.expectedResp.subResponses), len(outs))

	for i, share := range outs[0].Shares {
		require.Equal(t, share.Name, testReqResp.expectedResp.subResponses[0].Shares[i].Name)
		require.Equal(t, share.Price, testReqResp.expectedResp.subResponses[0].Shares[i].Price)
	}
	for i, share := range outs[1].Shares {
		require.Equal(t, share.Name, testReqResp.expectedResp.subResponses[0].Shares[i].Name)
		require.Equal(t, share.Price, testReqResp.expectedResp.subResponses[0].Shares[i].Price)
	}

	closer()
}

func TestValidationID(t *testing.T) {
	s := new(mocks.PriceServiceService)
	h := NewHandler(s, validate)

	testID := uuid.New()
	parsedID, err := h.ValidationID(context.Background(), testID.String())
	require.NoError(t, err)
	require.Equal(t, testID, parsedID)
}
