package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/cyberark/conjur-k8s-csi-provider/pkg/conjur"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

type mockConjurClient struct {
	resp map[string][]byte
	err  error
}

func (c *mockConjurClient) GetSecrets(jwt string, secretIds []string) (map[string][]byte, error) {
	return c.resp, c.err
}

func TestMount(t *testing.T) {
	testCases := []struct {
		description        string
		req                *v1alpha1.MountRequest
		conjurFactory      conjur.ClientFactory
		getAnnotationsFunc k8s.GetPodAnnotationsFunc
		assertions         func(*testing.T, *v1alpha1.MountResponse, error)
	}{
		{
			description: "throws error decoding invalid attributes",
			req: &v1alpha1.MountRequest{
				Attributes: "}invalid,json{",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Failed to unmarshal attributes")
			},
		},
		{
			description: "throws error decoding invalid attributes",
			req: &v1alpha1.MountRequest{
				Attributes: `{"conjur.org/configurationVersion": "0.3.0"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Unsupported configuration version")
			},
		},
		{
			description: "throws error decoding invalid serviceaccount tokens attribute",
			req: &v1alpha1.MountRequest{
				Attributes: "{\"csi.storage.k8s.io/serviceAccount.tokens\":\"invalid,json\"}",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), fmt.Sprintf("Failed to unmarshal attribute %q", saTokensKey))
			},
		},
		{
			description: "throws error when missing serviceaccount token for audience",
			req: &v1alpha1.MountRequest{
				Attributes: "{\"csi.storage.k8s.io/serviceAccount.tokens\":\"{}\"}",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Missing serviceaccount token for audience \"conjur\"")
			},
		},
		{
			description: "throws error when Conjur config not included",
			req: &v1alpha1.MountRequest{
				Attributes: `{"applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), `Missing required Conjur config attributes: ["account" "identity" "sslCertificate"]`)
			},
		},
		{
			description: "throws error when secrets attribute not included or empty (v0.1.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"conjur.org/configurationVersion":"0.1.0","secrets":"","sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Attribute \"secrets\" missing or empty")
			},
		},
		{
			description: "throws error when secrets attribute improperly formatted (v0.1.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"conjur.org/configurationVersion":"0.1.0","secrets":"invalid","sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Failed to unmarshal secrets spec")
			},
		},
		{
			description: "throws error when secrets attribute not included or empty (v0.2.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			getAnnotationsFunc: func(namespace string, podName string) (map[string]string, error) {
				return map[string]string{
					"conjur.org/secrets":    "",
					"some-other-annotation": "some-value",
				}, nil
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "CKCP034 Annotation \"conjur.org/secrets\" missing or empty")
			},
		},
		{
			description: "throws error when secrets attribute improperly formatted (v0.2.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			getAnnotationsFunc: func(namespace string, podName string) (map[string]string, error) {
				return map[string]string{
					"conjur.org/secrets":    "invalid",
					"some-other-annotation": "some-value",
				}, nil
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Failed to unmarshal secrets spec")
			},
		},
		{
			description: "throws error decoding invalid file permissions",
			req: &v1alpha1.MountRequest{
				Attributes: `{"sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "abc",
			},
			getAnnotationsFunc: func(namespace string, podName string) (map[string]string, error) {
				return map[string]string{
					"conjur.org/secrets": "- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n",
				}, nil
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Failed to unmarshal file permissions")
			},
		},
		{
			description: "throws error when conjur client fails",
			req: &v1alpha1.MountRequest{
				Attributes: `{"sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "777",
				TargetPath: "/some/path",
			},
			getAnnotationsFunc: func(namespace string, podName string) (map[string]string, error) {
				return map[string]string{
					"conjur.org/secrets": "- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n",
				}, nil
			},
			conjurFactory: func(baseURL, authnID, account, identity, sslCert string) conjur.Client {
				return &mockConjurClient{
					resp: nil,
					err:  errors.New("Conjur error getting secrets"),
				}
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Failed to get Conjur secrets")
			},
		},
		{
			description: "happy path (v0.1.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"conjur.org/configurationVersion":"0.1.0","secrets":"- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n","sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "777",
				TargetPath: "/some/path",
			},
			conjurFactory: func(baseURL, authnID, account, identity, sslCert string) conjur.Client {
				return &mockConjurClient{
					resp: map[string][]byte{
						"conjur/path/A": []byte("contentA"),
						"conjur/path/B": []byte("contentB"),
					},
					err: nil,
				}
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, err)

				assert.Len(t, resp.ObjectVersion, 2)
				assert.Len(t, resp.Files, 2)

				assert.Contains(t, resp.ObjectVersion, &v1alpha1.ObjectVersion{
					Id:      "conjur/path/A",
					Version: "1",
				})
				assert.Contains(t, resp.ObjectVersion, &v1alpha1.ObjectVersion{
					Id:      "conjur/path/B",
					Version: "1",
				})
				assert.Contains(t, resp.Files, &v1alpha1.File{
					Path:     "file/path/A",
					Mode:     int32(777),
					Contents: []byte("contentA"),
				})
				assert.Contains(t, resp.Files, &v1alpha1.File{
					Path:     "file/path/B",
					Mode:     int32(777),
					Contents: []byte("contentB"),
				})
			},
		},
		{
			description: "happy path (v0.2.0)",
			req: &v1alpha1.MountRequest{
				Attributes: `{"sslCertificate":"certificate content","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "777",
				TargetPath: "/some/path",
			},
			conjurFactory: func(baseURL, authnID, account, identity, sslCert string) conjur.Client {
				return &mockConjurClient{
					resp: map[string][]byte{
						"conjur/path/A": []byte("contentA"),
						"conjur/path/B": []byte("contentB"),
					},
					err: nil,
				}
			},
			getAnnotationsFunc: func(namespace string, podName string) (map[string]string, error) {
				return map[string]string{
					"conjur.org/secrets":    "- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n",
					"some-other-annotation": "some-value",
				}, nil
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, err)

				assert.Len(t, resp.ObjectVersion, 2)
				assert.Len(t, resp.Files, 2)

				assert.Contains(t, resp.ObjectVersion, &v1alpha1.ObjectVersion{
					Id:      "conjur/path/A",
					Version: "1",
				})
				assert.Contains(t, resp.ObjectVersion, &v1alpha1.ObjectVersion{
					Id:      "conjur/path/B",
					Version: "1",
				})
				assert.Contains(t, resp.Files, &v1alpha1.File{
					Path:     "file/path/A",
					Mode:     int32(777),
					Contents: []byte("contentA"),
				})
				assert.Contains(t, resp.Files, &v1alpha1.File{
					Path:     "file/path/B",
					Mode:     int32(777),
					Contents: []byte("contentB"),
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resp, err := mountWithDeps(context.TODO(), tc.req, tc.conjurFactory, tc.getAnnotationsFunc)
			tc.assertions(t, resp, err)
		})
	}
}

func TestVersion(t *testing.T) {
	testCases := []struct {
		description string
		req         *v1alpha1.VersionRequest
		assertions  func(*testing.T, *v1alpha1.VersionResponse, error)
	}{
		{
			description: "response csi driver version echos request",
			req: &v1alpha1.VersionRequest{
				Version: "0.0.test",
			},
			assertions: func(t *testing.T, resp *v1alpha1.VersionResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "0.0.test", resp.Version)
			},
		},
		{
			description: "response includes provider details",
			req: &v1alpha1.VersionRequest{
				Version: "0.0.test",
			},
			assertions: func(t *testing.T, resp *v1alpha1.VersionResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "conjur", resp.RuntimeName)
				expectedVersion, err := os.ReadFile("/conjur-k8s-csi-provider/VERSION")
				if err != nil {
					expectedVersion = []byte("0.0")
				}
				assert.Equal(t, string(expectedVersion), resp.RuntimeVersion)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resp, err := Version(context.TODO(), tc.req)
			tc.assertions(t, resp, err)
		})
	}
}
