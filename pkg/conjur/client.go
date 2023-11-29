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

// Client holds the configuration needed to communicate with Conjur
//
// Example:
//
//	client := Client{
//	    BaseURL:  "https://conjur-conjur-oss.conjur.svc.cluster.local",
//	    AuthnID:  "authn-jwt/kube",
//	    Account:  "default",
//	    Identity: "host/host1",
//	}
type Client struct {
	BaseURL  string
	AuthnID  string
	Account  string
	Identity string
}

func NewClient(baseURL, authnID, account, identity string) *Client {
	return &Client{
		BaseURL:  baseURL,
		AuthnID:  authnID,
		Account:  account,
		Identity: identity,
	}
}

func (c *Client) authenticate(token string) ([]byte, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, c.AuthnID, c.Account, url.PathEscape(c.Identity), "authenticate")

	data := url.Values{}
	data.Set("jwt", token)

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

func (c *Client) GetSecrets(token string, secretIds []string) (map[string][]byte, error) {
	authenticatedToken, err := c.authenticate(token)
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
