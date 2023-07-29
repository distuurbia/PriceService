package handler

// import (
// 	"testing"

// 	"github.com/distuurbia/PriceService/internal/handler/mocks"
// 	"github.com/distuurbia/PriceService/proto_services"
// 	"github.com/stretchr/testify/mock"
// )

// func TestReadFromStream(t *testing.T) {
// 	priceServiceSrv := new(mocks.PriceServiceService)
// 	priceServiceSrv.On("ReadFromStream", mock.Anything).
// 		Return(testShares, nil).
// 		Once()
// 	GRPCHandl := NewGRPCHandler(priceServiceSrv)
// 	var stream proto_services.PriceServiceService_ReadFromStreamServer
// 	err := GRPCHandl.ReadFromStream(&proto_services.ReadFromStreamRequest{}, stream)

// 	servCar.AssertExpectations(t)
// }
