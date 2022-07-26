package networks

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/pkg/errors"
)

type SubstrateClient struct {
	api      *gsrpc.SubstrateAPI
	metadata *types.Metadata
}

func NewSubstrateClient(networkConfig *NetworkConfig) (*SubstrateClient, error) {
	rpc := networkConfig.Hosts[0].URL
	api, err := getApi(rpc)
	if err != nil {
		return nil, errors.Wrap(err, "error occurred while creating substrate api")
	}
	metadata, err := getMetadata(api)
	if err != nil {
		return nil, errors.Wrap(err, "error occurred while fetching substrate rpc metadata")
	}
	return &SubstrateClient{
		api:      api,
		metadata: metadata,
	}, nil
}

func getApi(rpcAddress string) (*gsrpc.SubstrateAPI, error) {
	return gsrpc.NewSubstrateAPI(rpcAddress)
}

func getMetadata(api *gsrpc.SubstrateAPI) (*types.Metadata, error) {
	return api.RPC.State.GetMetadataLatest()
}
