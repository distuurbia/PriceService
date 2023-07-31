package handler

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/distuurbia/PriceService/internal/handler/mocks"
	protocol "github.com/distuurbia/PriceService/protocol/price"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var testShares = []*protocol.Share{
	{Name: "Apple", Price: 250},
	{Name: "Tesla", Price: 1000},
}

func Server(ctx context.Context, s *mocks.PriceServiceService) (psClient protocol.PriceServiceServiceClient, clientCloser func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)
	baseServer := grpc.NewServer()
	protocol.RegisterPriceServiceServiceServer(baseServer, NewHandler(s))
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			logrus.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			logrus.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := protocol.NewPriceServiceServiceClient(conn)

	return client, closer
}

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}
