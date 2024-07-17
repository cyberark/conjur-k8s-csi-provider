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

// Client is an interface to Conjur's API functions required by our CSI Provider.
type Client interface {
	GetSecrets(jwt string, secretIds []string) (map[string][]byte, error)
}

// Config holds the configuration needed to communicate with Conjur and
// implements the Client interface.
//
// Example:
//
//	config := Config{
//	    BaseURL:  "https://conjur-conjur-oss.conjur.svc.cluster.local",
//	    AuthnID:  "authn-jwt/kube",
//	    Account:  "default",
//	    Identity: "host/host1",
//	}
type Config struct {
	BaseURL  string
	AuthnID  string
	Account  string
	Identity string
	SSLCert  string
}

// NewClient returns a new Conjur client.
func NewClient(baseURL, authnID, account, identity, sslCert string) Client {
	return &Config{
		BaseURL:  baseURL,
		AuthnID:  authnID,
		Account:  account,
		Identity: identity,
		SSLCert:  sslCert,
	}
}

// authenticate creates an authenticated Conjur client using the provided JWT.
func (c *Config) authenticate(jwt string) (*conjurapi.Client, error) {
	// Parse the service ID if needed
	serviceId := c.AuthnID
	if strings.Contains(c.AuthnID, "authn-jwt/") {
		serviceId = strings.Split(c.AuthnID, "authn-jwt/")[1]
	}

	config := conjurapi.Config{
		Account:      c.Account,
		ApplianceURL: c.BaseURL,
		SSLCert:      c.SSLCert,
		AuthnType:    "jwt",
		ServiceID:    serviceId,
		JWTHostID:    c.Identity,
		JWTContent:   jwt,
	}

	err := config.Validate()
	if err != nil {
		log.Error(logmessages.CKCP030, err)
		return nil, fmt.Errorf(logmessages.CKCP030, err)
	}

	conjur, err := conjurapi.NewClientFromJwt(config)
	if err != nil {
		log.Error(logmessages.CKCP030, err)
		return nil, fmt.Errorf(logmessages.CKCP030, err)
	}

	return conjur, nil
}

// GetSecrets authenticates with Conjur using the provided JWT and returns
// requested secret data.
func (c *Config) GetSecrets(jwt string, secretIds []string) (map[string][]byte, error) {
	authenticatedClient, err := c.authenticate(jwt)
	if err != nil {
		log.Error(logmessages.CKCP031, err)
		return nil, fmt.Errorf(logmessages.CKCP031, err)
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
