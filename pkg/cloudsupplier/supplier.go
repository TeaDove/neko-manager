package cloudsupplier

import (
	"context"

	"github.com/pkg/errors"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"

	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
)

type Supplier struct {
	sdk        *ycsdk.SDK
	computeSDK computesdk.InstanceClient
	vpcSDK     vpcsdk.SubnetClient

	folderID string
	zone     string
	subnetID string
}

func New(ctx context.Context, sdk *ycsdk.SDK, folderID string) (*Supplier, error) {
	r := &Supplier{sdk: sdk, computeSDK: computesdk.NewInstanceClient(sdk), folderID: folderID}
	r.zone = "ru-central1-b"

	var err error
	r.subnetID, err = r.getDefaultNetworks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get default subnets")
	}

	return r, nil
}

func (r *Supplier) getDefaultNetworks(ctx context.Context) (string, error) {
	networkSDK := vpcsdk.NewNetworkClient(r.sdk)

	listResp, err := networkSDK.List(ctx, &vpc.ListNetworksRequest{
		FolderId: r.folderID,
		Filter:   "Network.name=default",
	})
	if err != nil {
		return "", errors.Wrap(err, "list networks")
	}

	if len(listResp.Networks) == 0 {
		return "", errors.New("no default network")
	}

	listSubnetsResp, err := networkSDK.ListSubnets(ctx, &vpc.ListNetworkSubnetsRequest{
		NetworkId: listResp.Networks[0].Id,
	})
	if err != nil {
		return "", errors.Wrap(err, "list subnets")
	}

	for _, subnet := range listSubnetsResp.Subnets {
		if subnet.ZoneId == r.zone {
			return subnet.Id, nil
		}
	}

	return "", errors.New("no default network")
}
