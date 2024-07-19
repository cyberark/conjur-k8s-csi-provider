package conjur

import (
	"fmt"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
)

// ClientFactory returns an implementation of the Client interface given the
// proper configuration values.
type ClientFactory func(baseURL, authnID, account, identity, sslCert string) Client

// Client is an interface to functions required by our CSI Provider.
type Client interface {
	GetSecrets(jwt string, secretIds []string) (map[string][]byte, error)
}

// ConjurClient interface for the methods we use from conjurapi.Client
type ConjurClient interface {
	RetrieveBatchSecretsSafe([]string) (map[string][]byte, error)
}

// Config holds the configuration needed to communicate with Conjur and
// implements the Client interface.
type Config struct {
	BaseURL       string
	AuthnID       string
	Account       string
	Identity      string
	SSLCert       string
	clientFactory func(conjurapi.Config) (ConjurClient, error)
}

// NewClient returns a new Conjur client.
func NewClient(baseURL, authnID, account, identity, sslCert string) Client {
	return &Config{
		BaseURL:       baseURL,
		AuthnID:       authnID,
		Account:       account,
		Identity:      identity,
		SSLCert:       sslCert,
		clientFactory: defaultClientFactory,
	}
}

func defaultClientFactory(config conjurapi.Config) (ConjurClient, error) {
	return conjurapi.NewClientFromJwt(config)
}

// GetSecrets authenticates with Conjur using the provided JWT and returns
// requested secret data.
func (c *Config) GetSecrets(jwt string, secretIds []string) (map[string][]byte, error) {
	serviceID := c.AuthnID
	if strings.Contains(c.AuthnID, "authn-jwt/") {
		serviceID = strings.Split(c.AuthnID, "authn-jwt/")[1]
	}

	config := conjurapi.Config{
		Account:      c.Account,
		ApplianceURL: c.BaseURL,
		SSLCert:      c.SSLCert,
		AuthnType:    "jwt",
		ServiceID:    serviceID,
		JWTHostID:    c.Identity,
		JWTContent:   jwt,
	}

	if err := config.Validate(); err != nil {
		log.Error(logmessages.CKCP030, err)
		return nil, fmt.Errorf(logmessages.CKCP030, err)
	}

	authenticatedClient, err := c.clientFactory(config)
	if err != nil {
		log.Error(logmessages.CKCP030, err)
		return nil, fmt.Errorf(logmessages.CKCP030, err)
	}

	secretValuesByID := map[string][]byte{}
	secretValuesByFullID, err := authenticatedClient.RetrieveBatchSecretsSafe(secretIds)
	if err != nil {
		log.Error(logmessages.CKCP031, err)
		return nil, fmt.Errorf(logmessages.CKCP031, err)
	}

	prefix := fmt.Sprintf("%s:variable:", c.Account)
	for k, v := range secretValuesByFullID {
		secretValuesByID[strings.TrimPrefix(k, prefix)] = v
	}
	return secretValuesByID, nil
}
