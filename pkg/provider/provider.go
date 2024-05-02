package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/conjur"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/k8s"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const providerName = "conjur"
const saTokensKey = "csi.storage.k8s.io/serviceAccount.tokens"
const podNameKey = "csi.storage.k8s.io/pod.name"
const podNamespaceKey = "csi.storage.k8s.io/pod.namespace"
const configurationVersionKey = "conjur.org/configurationVersion"
const secretsAnnotationKey = "conjur.org/secrets"

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
	return mountWithDeps(ctx, req, conjur.NewClient, k8s.GetPodAnnotations)
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
	getAnnotationsFunc k8s.GetPodAnnotationsFunc,
) (*v1alpha1.MountResponse, error) {
	cfg, err := NewConfig(req, getAnnotationsFunc)
	if err != nil {
		log.Error(logmessages.CKCP013, err)
		return nil, fmt.Errorf(logmessages.CKCP013, err)
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
		log.Error(logmessages.CKCP016, err)
		return nil, fmt.Errorf(logmessages.CKCP016, err)
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
		log.Error(logmessages.CKCP017, err)
		return nil, fmt.Errorf(logmessages.CKCP017, err)
	}

	return attributes, nil
}

func NewConfig(req *v1alpha1.MountRequest, getAnnotationsFunc k8s.GetPodAnnotationsFunc) (*Config, error) {
	var tokens map[string]map[string]string
	var token string
	var secretsStr string
	var secrets map[string]string
	var permissions os.FileMode
	var configVersion *version.Version
	var err error

	attributes, err := parseRequestAttributes(req)
	if err != nil {
		log.Error(logmessages.CKCP032, err)
		return nil, fmt.Errorf(logmessages.CKCP032, err)
	}

	configVersionStr := attributes[configurationVersionKey]
	switch configVersionStr {
	case "0.1.0", "0.2.0":
		log.Info(logmessages.CKCP040, configVersionStr)
		configVersion, _ = version.NewVersion(configVersionStr)
	case "":
		log.Info(logmessages.CKCP041, "0.2.0")
		configVersion, _ = version.NewVersion("0.2.0")
	default:
		log.Error(logmessages.CKCP006, configVersionStr)
		return nil, fmt.Errorf(logmessages.CKCP006, configVersionStr)
	}

	err = json.Unmarshal([]byte(attributes[saTokensKey]), &tokens)
	if err != nil {
		log.Error(logmessages.CKCP007, saTokensKey, err)
		return nil, fmt.Errorf(logmessages.CKCP007, saTokensKey, err)
	}

	token = tokens[providerName]["token"]
	if token == "" {
		log.Error(logmessages.CKCP008, providerName)
		return nil, fmt.Errorf(logmessages.CKCP008, providerName)
	}

	missingKeys := []string{}
	for _, key := range []string{"account", "applianceUrl", "authnId", "identity", "sslCertificate"} {
		if attributes[key] == "" {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 {
		log.Error(logmessages.CKCP009, missingKeys)
		return nil, fmt.Errorf(logmessages.CKCP009, missingKeys)
	}

	// Starting with configurationVersion 0.2.0, the 'secrets' attribute is
	// retrieved from the 'conjur.org/secrets' annotation of the application pod.
	// Prior to 0.2.0, the 'secrets' attribute is expected to be provided in the
	// MountRequest attributes from the SecretProviderClass params.
	annotationVersion, _ := version.NewVersion("0.2.0")
	if configVersion.GreaterThanOrEqual(annotationVersion) {
		secretsStr, err = retrievePodAnnotationSecrets(attributes, getAnnotationsFunc)
		if err != nil {
			log.Error(logmessages.CKCP035, err)
			return nil, fmt.Errorf(logmessages.CKCP035, err)
		}
	} else {
		secretsStr = attributes["secrets"]
	}

	if secretsStr == "" {
		log.Error(logmessages.CKCP010, "secrets")
		return nil, fmt.Errorf(logmessages.CKCP010, "secrets")
	}

	secrets, err = parseSecrets(secretsStr)
	if err != nil {
		log.Error(logmessages.CKCP011, err)
		return nil, fmt.Errorf(logmessages.CKCP011, err)
	}

	err = json.Unmarshal([]byte(req.GetPermission()), &permissions)
	if err != nil {
		log.Error(logmessages.CKCP012, err)
		return nil, fmt.Errorf(logmessages.CKCP012, err)
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
		log.Error(logmessages.CKCP033, err)
		return nil, fmt.Errorf(logmessages.CKCP033, err)
	}

	returned := make(map[string]string, len(intermediate))
	for _, i := range intermediate {
		for k, v := range i {
			returned[v] = k
		}
	}

	return returned, nil
}

// retrievePodAnnotationSecrets retrieves the annotation 'conjur.org/secrets'
// from the pod that is associated with a given MountRequest. The annotation
// value is assumed to match the YAML format expected by parseSecrets.
func retrievePodAnnotationSecrets(attributes map[string]string, getAnnotationsFunc k8s.GetPodAnnotationsFunc) (string, error) {
	annotations, err := getAnnotationsFunc(attributes[podNamespaceKey], attributes[podNameKey])
	if err != nil {
		log.Error(logmessages.CKCP033, err)
		return "", fmt.Errorf(logmessages.CKCP033, err)
	}

	if annotations[secretsAnnotationKey] == "" {
		log.Error(logmessages.CKCP034, secretsAnnotationKey)
		return "", fmt.Errorf(logmessages.CKCP034, secretsAnnotationKey)
	}

	return annotations[secretsAnnotationKey], nil
}
