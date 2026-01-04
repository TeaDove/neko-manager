package cloudsupplier

import (
	bytes "bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Supplier) ComputeCreateWaited(
	ctx context.Context,
	id string,
	name string,
	createdBy string,
	sessionAPIToken string,
	resourcesSpec *compute.ResourcesSpec,
) (string, error) {
	latestImageResp, err := computesdk.NewImageClient(r.sdk).
		GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
			FolderId: "yc.container-solution",
			Family:   "container-optimized-image",
		})
	if err != nil {
		return "", errors.Wrap(err, "fetch latest image")
	}

	var dockerCompose bytes.Buffer

	err = r.nekoDockerComposeTemplate.Execute(&dockerCompose, map[string]any{
		"sessionAPIToken": sessionAPIToken,
		"pathPrefix":      fmt.Sprintf("/%s", id),
	})
	if err != nil {
		return "", errors.Wrap(err, "execute neko template")
	}

	description := fmt.Sprintf("automated neko, createdBy=%s", createdBy)

	operation, err := r.computeSDK.Create(ctx, &compute.CreateInstanceRequest{
		FolderId:         r.FolderID,
		ZoneId:           r.zone,
		Name:             name,
		Description:      description,
		SchedulingPolicy: &compute.SchedulingPolicy{Preemptible: true},
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{{
			Index:    "0",
			SubnetId: r.subnetID,
			PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
				OneToOneNatSpec: &compute.OneToOneNatSpec{IpVersion: compute.IpVersion_IPV4},
			},
		}},
		Labels: map[string]string{"project": "neko-manager", "created-by": strings.ToLower(createdBy), "id": id},
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
				Name:        name,
				Description: description,
				TypeId:      "network-ssd",
				Size:        1024 * 1024 * 1024 * 30,
				Source:      &compute.AttachedDiskSpec_DiskSpec_ImageId{ImageId: latestImageResp.GetId()},
			}},
		},
		PlatformId:    "standard-v4a", // standard-v3
		ResourcesSpec: resourcesSpec,
		Hostname:      name,
		Metadata: map[string]string{
			"docker-compose": dockerCompose.String(),
			"user-data": fmt.Sprintf(
				`#cloud-config
datasource:
 Ec2:
  strict_id: false
ssh_pwauth: no
users:
- name: %s
  sudo: ALL=(ALL) NOPASSWD:ALL
  shell: /bin/bash
  ssh_authorized_keys:
  - %s`,
				r.sshUsername,
				r.sshPublicKey,
			),
			"ssh-keys": fmt.Sprintf("%s:%s", r.sshUsername, r.sshPublicKey),
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "create operation")
	}

	cloudInstanceID := operation.Metadata().GetInstanceId()

	zerolog.Ctx(ctx).Info().
		Str("cloud_instance_id", cloudInstanceID).
		Str("operation_id", operation.ID()).
		Msg("cloud.instance.creating")

	_, err = operation.Wait(ctx)
	if err != nil {
		return "", errors.Wrap(err, "wait operation")
	}

	return cloudInstanceID, nil
}

func (r *Supplier) ComputeGet(ctx context.Context, name string) (*compute.Instance, error) {
	resp, err := r.computeSDK.List(ctx, &compute.ListInstancesRequest{
		FolderId: r.FolderID,
		Filter:   fmt.Sprintf(`name="%s"`, name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list instances")
	}

	if len(resp.GetInstances()) == 0 {
		return nil, status.New(codes.NotFound, fmt.Sprintf("instance not found: %s", name)).Err()
	}

	return resp.GetInstances()[0], nil
}

func (r *Supplier) ComputeList(ctx context.Context, name string) ([]*compute.Instance, error) {
	resp, err := r.computeSDK.List(ctx, &compute.ListInstancesRequest{
		FolderId: r.FolderID,
		Filter:   fmt.Sprintf(`name CONTAINS "%s"`, name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list instances")
	}

	return resp.GetInstances(), nil
}

func (r *Supplier) ComputeDeleteWaited(ctx context.Context, cloudId string) error {
	operation, err := r.computeSDK.Delete(ctx, &compute.DeleteInstanceRequest{InstanceId: cloudId})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			zerolog.Ctx(ctx).
				Warn().
				Str("cloud_instance_id", cloudId).
				Msg("instance.not.found")

			return nil
		}

		return errors.Wrap(err, "delete")
	}

	zerolog.Ctx(ctx).Info().
		Str("cloud_instance_id", cloudId).
		Str("operation_id", operation.ID()).
		Msg("cloud.instance.deleting")

	_, err = operation.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "wait")
	}

	zerolog.Ctx(ctx).Info().
		Str("cloud_instance_id", cloudId).
		Msg("cloud.instance.deleted")

	return nil
}

func (r *Supplier) ComputeRestartWaited(ctx context.Context, cloudId string) error {
	operation, err := r.computeSDK.Restart(ctx, &compute.RestartInstanceRequest{InstanceId: cloudId})
	if err != nil {
		return errors.Wrap(err, "restart")
	}

	_, err = operation.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "wait")
	}

	zerolog.Ctx(ctx).Info().
		Str("cloud_instance_id", cloudId).
		Msg("cloud.instance.restarted")

	return nil
}
