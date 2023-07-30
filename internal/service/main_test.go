package service

import (
	"os"
	"testing"

	"github.com/distuurbia/PriceService/internal/model"
	"github.com/google/uuid"
)

var (
	testSubID          = uuid.New()
	testSelectedShares = []string{
		"Apple",
		"Tesla",
	}
	testShares = []*model.Share{
		{Name: "Apple", Price: 500},
		{Name: "Tesla", Price: 600},
		{Name: "Google", Price: 1000},
	}
)

func TestMain(m *testing.M){
	exitCode := m.Run()
	os.Exit(exitCode)
}