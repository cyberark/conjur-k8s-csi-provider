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
const saTokensKey = "csi.storage.k8s.io/serviceAccount.tokens"
const configurationVersionKey = "conjur.org/configurationVersion"

// Config contains information parses from a Mount request that is required for
// authenticating with Conjur and retrieving secrets.
type Config struct {
	// Custom attributes attached to a given MountRequest
	attributes map[string]string
	// ServiceAccount JWT token used to authenticate to Conjur
	token string
	// Desired permissions on generated secret files
	permissions os.FileMode
	// Secrets spec relating Conjur secret IDs to file paths
	secrets map[string]string
}

// Mount implements a volume mount operation in the Conjur provider
func Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	return mountWithDeps(
		ctx, req, conjur.NewClient,
	)
}

// Version returns Conjur provider runtime details
func Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return &v1alpha1.VersionResponse{
		Version:        req.GetVersion(),
		RuntimeName:    providerName,
		RuntimeVersion: ProviderVersion,
	}, nil
}

func mountWithDeps(
	ctx context.Context,
	req *v1alpha1.MountRequest,
	conjurFactory conjur.ClientFactory,
) (*v1alpha1.MountResponse, error) {
	cfg, err := NewConfig(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration from mount request parameters: %w", err)
	}

	secretIDs := []string{}
	for secretID, _ := range cfg.secrets {
		secretIDs = append(secretIDs, secretID)
	}
	conjClient := conjurFactory(
		cfg.attributes["applianceUrl"],
		cfg.attributes["authnId"],
		cfg.attributes["account"],
		cfg.attributes["identity"],
		cfg.attributes["sslCertificate"],
	)
	secrets, err := conjClient.GetSecrets(cfg.token, secretIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get Conjur secrets: %w", err)
	}

	objectVersion := []*v1alpha1.ObjectVersion{}
	files := []*v1alpha1.File{}

	for secretID, value := range secrets {
		objectVersion = append(objectVersion, &v1alpha1.ObjectVersion{
			Id:      secretID,
			Version: "1",
		})
		files = append(files, &v1alpha1.File{
			Path:     cfg.secrets[secretID],
			Mode:     int32(cfg.permissions),
			Contents: value,
		})
	}

	return &v1alpha1.MountResponse{
		ObjectVersion: objectVersion,
		Files:         files,
	}, nil
}

func parseRequestAttributes(req *v1alpha1.MountRequest) (map[string]string, error) {
	var attributes map[string]string

	err := json.Unmarshal([]byte(req.GetAttributes()), &attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	return attributes, nil
}

func NewConfig(req *v1alpha1.MountRequest) (*Config, error) {
	var tokens map[string]map[string]string
	var token string
	var secretsStr string
	var secrets map[string]string
	var permissions os.FileMode
	var err error

	attributes, err := parseRequestAttributes(req)
	if err != nil {
		return nil, err
	}

	configVersion := attributes[configurationVersionKey]
	if len(configVersion) > 0 && configVersion != "0.1.0" {
		return nil, fmt.Errorf("unsupported configuration version: %q", configVersion)
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
	for _, key := range []string{"account", "applianceUrl", "authnId", "identity", "sslCertificate"} {
		if attributes[key] == "" {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 {
		return nil, fmt.Errorf("missing required Conjur config attributes: %q", missingKeys)
	}

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
		attributes:  attributes,
		token:       token,
		permissions: permissions,
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
			returned[v] = k
		}
	}

	return returned, nil
}
