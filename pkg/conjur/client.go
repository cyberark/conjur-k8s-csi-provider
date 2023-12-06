package conjur

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// ClientFactory returns an implementation of the Client interface given the
// proper configuration values.
type ClientFactory func(baseURL, authnID, account, identity string) Client

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
}

// NewClient returns a new Conjur client.
func NewClient(baseURL, authnID, account, identity string) Client {
	return &Config{
		BaseURL:  baseURL,
		AuthnID:  authnID,
		Account:  account,
		Identity: identity,
	}
}

func (c *Config) authenticate(jwt string) ([]byte, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, c.AuthnID, c.Account, url.PathEscape(c.Identity), "authenticate")

	data := url.Values{}
	data.Set("jwt", jwt)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// GetSecrets authenticates with Conjur using the provided JWT and returns
// requested secret data.
func (c *Config) GetSecrets(jwt string, secretIds []string) (map[string][]byte, error) {
	authenticatedToken, err := c.authenticate(jwt)
	if err != nil {
		return nil, err
	}

	conjur, err := conjurapi.NewClientFromToken(conjurapi.Config{Account: c.Account, ApplianceURL: c.BaseURL}, string(authenticatedToken))
	if err != nil {
		return nil, err
	}
	conjur.SetHttpClient(&http.Client{})

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
