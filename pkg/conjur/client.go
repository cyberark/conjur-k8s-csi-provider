package conjur

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *Config) authenticate(jwt string) ([]byte, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL = requestURL.JoinPath(
		c.AuthnID,
		c.Account,
		url.PathEscape(c.Identity),
		"authenticate",
	)

	data := url.Values{}
	data.Set("jwt", jwt)

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM([]byte(c.SSLCert))
	if !ok {
		log.Error(logmessages.CKCP014)
		return nil, fmt.Errorf(logmessages.CKCP014)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest(
		"POST",
		requestURL.String(),
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Error(logmessages.CKCP015, resp.StatusCode)
		return nil, fmt.Errorf(logmessages.CKCP015, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetSecrets authenticates with Conjur using the provided JWT and returns
// requested secret data.
func (c *Config) GetSecrets(jwt string, secretIds []string) (map[string][]byte, error) {
	authenticatedToken, err := c.authenticate(jwt)
	if err != nil {
		return nil, err
	}

	conjur, err := conjurapi.NewClientFromToken(
		conjurapi.Config{
			Account:      c.Account,
			ApplianceURL: c.BaseURL,
			SSLCert:      c.SSLCert,
		}, string(authenticatedToken))
	if err != nil {
		return nil, err
	}

	secretValuesByID := map[string][]byte{}
	secretValuesByFullID, err := conjur.RetrieveBatchSecretsSafe(secretIds)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s:variable:", c.Account)
	for k, v := range secretValuesByFullID {
		secretValuesByID[strings.TrimPrefix(k, prefix)] = v
	}
	return secretValuesByID, nil
}
