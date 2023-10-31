package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

func TestMount(t *testing.T) {
	testCases := []struct {
		description string
		req         *v1alpha1.MountRequest
		assertions  func(*testing.T, *v1alpha1.MountResponse, error)
	}{
		{
			description: "throws error decoding invalid attributes",
			req: &v1alpha1.MountRequest{
				Attributes: "}invalid,json{",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "failed to unmarshal attributes")
			},
		},
		{
			description: "throws error decoding invalid serviceaccount tokens attribute",
			req: &v1alpha1.MountRequest{
				Attributes: "{\"csi.storage.k8s.io/serviceAccount.tokens\":\"invalid,json\"}",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), fmt.Sprintf("failed to unmarshal attribute %q", saTokensKey))
			},
		},
		{
			description: "throws error when missing serviceaccount token for audience",
			req: &v1alpha1.MountRequest{
				Attributes: "{\"csi.storage.k8s.io/serviceAccount.tokens\":\"{}\"}",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "missing serviceaccount token for audience \"conjur\"")
			},
		},
		{
			description: "throws error when Conjur config not included",
			req: &v1alpha1.MountRequest{
				Attributes: `{"applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), `missing required Conjur config attributes: ["account" "identity"]`)
			},
		},
		{
			description: "throws error when secrets attribute not included or empty",
			req: &v1alpha1.MountRequest{
				Attributes: `{"secrets":"","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "attribute \"secrets\" missing or empty")
			},
		},
		{
			description: "throws error when secrets attribute improperly formatted",
			req: &v1alpha1.MountRequest{
				Attributes: `{"secrets":"invalid","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "failed to unmarshal secrets spec")
			},
		},
		{
			description: "throws error decoding invalid file permissions",
			req: &v1alpha1.MountRequest{
				Attributes: `{"secrets":"- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "abc",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "failed to unmarshal file permissions")
			},
		},
		{
			description: "happy path",
			req: &v1alpha1.MountRequest{
				Attributes: `{"secrets":"- \"file/path/A\": \"conjur/path/A\"\n- \"file/path/B\": \"conjur/path/B\"\n","account":"default","applianceUrl":"https://my.conjur.com","authnId":"authn-jwt/instance","identity":"botApp","csi.storage.k8s.io/serviceAccount.tokens":"{\"conjur\":{\"token\":\"sometoken\",\"expirationTimestamp\":\"2123-01-01T01:01:01Z\"}}"}`,
				Permission: "777",
				TargetPath: "/some/path",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int32(777), resp.Files[0].Mode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resp, err := Mount(context.TODO(), tc.req)
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
			description: "response includes hardcoded provider details",
			req: &v1alpha1.VersionRequest{
				Version: "0.0.test",
			},
			assertions: func(t *testing.T, resp *v1alpha1.VersionResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "conjur", resp.RuntimeName)
				assert.Equal(t, "0.0.1", resp.RuntimeVersion)
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
