package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cyberark/conjur-k8s-csi-provider/pkg/conjur"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const providerName = "conjur"
const providerVersion = "0.0.1"
const saTokensKey = "csi.storage.k8s.io/serviceAccount.tokens"

// Config contains information parses from a Mount request that is required for
// authenticating with Conjur and retrieving secrets.
type Config struct {
	// ServiceAccount JWT token used to authenticate to Conjur
	token string
	// Desired permissions on generated secret files
	permissions os.FileMode
	// Conjur client for secret retrieval
	conjur *conjur.Client
	// Secrets spec relating file paths to Conjur secrets
	secrets map[string]string
}

// Mount implements a volume mount operation in the Conjur provider
func Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	cfg, err := parse(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mount request: %w", err)
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
				Mode:     int32(cfg.permissions),
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

func parse(req *v1alpha1.MountRequest) (*Config, error) {
	var attributes map[string]string
	var tokens map[string]map[string]string
	var token string
	var conjurClient *conjur.Client
	var secretsStr string
	var secrets map[string]string
	var permissions os.FileMode
	var err error

	err = json.Unmarshal([]byte(req.GetAttributes()), &attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	err = json.Unmarshal([]byte(attributes[saTokensKey]), &tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal attribute %q: %w", saTokensKey, err)
	}

	token = tokens[providerName]["token"]
	if token == "" {
		return nil, fmt.Errorf("missing serviceaccount token for audience %q", providerName)
	}

	missingKeys := []string{}
	for _, key := range []string{"account", "applianceUrl", "authnId", "identity"} {
		if attributes[key] == "" {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 {
		return nil, fmt.Errorf("missing required Conjur config attributes: %q", missingKeys)
	}
	conjurClient = conjur.NewClient(
		attributes["applianceUrl"],
		attributes["authnId"],
		attributes["account"],
		attributes["identity"],
	)

	secretsStr = attributes["secrets"]
	if secretsStr == "" {
		return nil, fmt.Errorf("attribute \"secrets\" missing or empty")
	}

	secrets, err = parseSecrets(secretsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets spec: %w", err)
	}

	err = json.Unmarshal([]byte(req.GetPermission()), &permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file permissions: %w", err)
	}

	return &Config{
		token:       token,
		permissions: permissions,
		conjur:      conjurClient,
		secrets:     secrets,
	}, nil
}

// parseSecrets expect the input string in the format:
//
// - "file/path/A": "conjur/path/A"
// - "file/path/B": "conjur/path/B"
//
// This format is recognized in YAML as a sequence of maps. Go's yaml.v3 package
// can parse the input string into a []map[string]string object, and we can
// transform the result into a map[string]string.
func parseSecrets(s string) (map[string]string, error) {
	var intermediate []map[string]string
	err := yaml.Unmarshal([]byte(s), &intermediate)
	if err != nil {
		return nil, err
	}

	returned := make(map[string]string, len(intermediate))
	for _, i := range intermediate {
		for k, v := range i {
			returned[k] = v
		}
	}

	return returned, nil
}
