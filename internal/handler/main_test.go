package handler

import (
	"os"
	"testing"

	"github.com/distuurbia/PriceService/internal/model"
)

var testShares []*model.Share

func TestMain(m *testing.M) {
	const berkshirePrice = 453066
	testShares = append(testShares, &model.Share{Name: "Berkshire Hathaway Inc.", Price: berkshirePrice})

	exitVal := m.Run()
	os.Exit(exitVal)
}
