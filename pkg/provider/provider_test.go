package provider

import (
	"context"
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
			description: "throws error decoding invalid secrets",
			req: &v1alpha1.MountRequest{
				Attributes: "{}",
				Secrets:    "}invalid,json{",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "failed to unmarshal secrets")
			},
		},
		{
			description: "throws error decoding invalid file permissions",
			req: &v1alpha1.MountRequest{
				Attributes: "{}",
				Secrets:    "{}",
				Permission: "abc",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "failed to unmarshal file permissions")
			},
		},
		{
			description: "throws error with empty target path",
			req: &v1alpha1.MountRequest{
				Attributes: "{}",
				Secrets:    "{}",
				Permission: "777",
				TargetPath: "",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "mount request missing target path")
			},
		},
		{
			description: "valid request begets response with desired file permissions",
			req: &v1alpha1.MountRequest{
				Attributes: "{}",
				Secrets:    "{}",
				Permission: "777",
				TargetPath: "/some/path",
			},
			assertions: func(t *testing.T, resp *v1alpha1.MountResponse, err error) {
				assert.NotNil(t, resp)
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
