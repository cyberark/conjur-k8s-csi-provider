package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const providerName = "conjur"
const providerVersion = "0.0.1"

// Mount implements a volume mount operation in the Conjur provider
func Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	var attributes, secrets map[string]string
	var permissions os.FileMode
	var path string
	var err error

	if err = json.Unmarshal([]byte(req.GetAttributes()), &attributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}
	if err = json.Unmarshal([]byte(req.GetSecrets()), &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets: %w", err)
	}
	if err = json.Unmarshal([]byte(req.GetPermission()), &permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file permissions: %w", err)
	}
	if path = req.GetTargetPath(); len(path) == 0 {
		return nil, fmt.Errorf("mount request missing target path")
	}

	return &v1alpha1.MountResponse{
		ObjectVersion: []*v1alpha1.ObjectVersion{
			{
				Id:      "someId",
				Version: "1",
			},
		},
		Files: []*v1alpha1.File{
			{
				Path:     "somePath.txt",
				Mode:     int32(permissions),
				Contents: []byte("someContent"),
			},
		},
	}, nil
}

// Version returns Conjur provider runtime details
func Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return &v1alpha1.VersionResponse{
		Version:        req.GetVersion(),
		RuntimeName:    providerName,
		RuntimeVersion: providerVersion,
	}, nil
}
