package cloudsupplier

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
)

func (r *Supplier) ComputeCreate(ctx context.Context, name string, createdBy string) error {
	latestImageResp, err := computesdk.NewImageClient(r.sdk).GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: "yc.container-solution",
		Family:   "container-optimized-image",
	})
	if err != nil {
		return errors.Wrap(err, "fetch latest image")
	}

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
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
				Name:        name,
				Description: description,
				TypeId:      "network-ssd",
				Size:        1024 * 1024 * 1024 * 30,
				Source:      &compute.AttachedDiskSpec_DiskSpec_ImageId{ImageId: latestImageResp.Id},
			}},
		},
		PlatformId: "standard-v4a", // standard-v3
		ResourcesSpec: &compute.ResourcesSpec{
			Memory:       1024 * 1024 * 1024 * 8,
			Cores:        8,
			CoreFraction: 100,
		},
		Hostname: name,
		Metadata: map[string]string{
			"docker-compose": nekoDockerCompose,
			"user-data":      fmt.Sprintf("#cloud-config\ndatasource:\n Ec2:\n  strict_id: false\nssh_pwauth: no\nusers:\n- name: %s\n  sudo: ALL=(ALL) NOPASSWD:ALL\n  shell: /bin/bash\n  ssh_authorized_keys:\n  - %s", r.sshUsername, r.sshPublicKey),
			"ssh-keys":       fmt.Sprintf("%s:%s", r.sshUsername, r.sshPublicKey),
		},
	})
	if err != nil {
		return errors.Wrap(err, "create operation")
	}

	zerolog.Ctx(ctx).Info().
		Str("operationID", operation.ID()).
		Msg("cloud.instance.creating")

	return nil
}

//go:embed neko-docker-compose.yaml
var nekoDockerCompose string
