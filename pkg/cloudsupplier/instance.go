package cloudsupplier

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/teadove/teasutils/utils/test_utils"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func (r *Supplier) ComputeCreate(ctx context.Context, name string, createdBy string) error {
	description := fmt.Sprintf("automated neko, createdBy=%s", createdBy)

	operation, err := r.computeSDK.Create(ctx, &compute.CreateInstanceRequest{
		FolderId:    r.folderID,
		ZoneId:      r.zone,
		Name:        name,
		Description: description,

		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{{
			Index:    "0",
			SubnetId: r.subnetID,
			PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
				OneToOneNatSpec: &compute.OneToOneNatSpec{IpVersion: compute.IpVersion_IPV4},
			},
		}},
		Labels: map[string]string{"project": "neko-manager"},
		//BootDiskSpec: &compute.AttachedDiskSpec{
		//	AutoDelete: true,
		//	Disk: &compute.AttachedDiskSpec_DiskSpec_{DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
		//		Name:        name,
		//		Description: description,
		//		TypeId:      "network-ssd",
		//		Size:        1024 * 1024 * 1024 * 10,
		//	}},
		//},
		PlatformId: "standard-v3",
		ResourcesSpec: &compute.ResourcesSpec{
			Memory:       1024 * 1024 * 1024 * 2,
			Cores:        2,
			CoreFraction: 50,
		},
		Hostname: name,
		Metadata: map[string]string{ // TODO add SSH connection
			"docker-compose": `services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:3.0.9"
    restart: "unless-stopped"
    shm_size: "4gb"
    ports:
      - "80:8080"
      - "59200-59300:59200-59300/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_MEMBER_PROVIDER: multiuser
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin

      NEKO_DESKTOP_SCREEN: '1920x1080@30'
      NEKO_WEBRTC_EPR: 59200-59300
      NEKO_FILE_TRANSFER_ENABLED: "true"

      NEKO_SERVER_METRICS: true

      NEKO_SESSION_IMPLICIT_HOSTING: true
      NEKO_SESSION_HEARTBEAT_INTERVAL: 120
      NEKO_SESSION_MERCIFUL_RECONNECT: true
`,
		},
	})
	if err != nil {
		return errors.Wrap(err, "create operation")
	}

	test_utils.Pprint(operation)

	return nil
}
