package connection

import (
	"rival/config"
	tigerbeetle_go "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

var itb tigerbeetle_go.Client

func NewTbClient() (tigerbeetle_go.Client, error) {
	if itb != nil {
		return itb, nil
	}
	config := config.GetConfig()
	tbAddress := config.Tb.Addr
	if len(tbAddress) == 0 {
		tbAddress = "3000"
	}
	client, err := tigerbeetle_go.NewClient(types.ToUint128(0), []string{tbAddress})
	if err != nil {
		return nil, err
	}
	itb = client
	return itb, nil
}
