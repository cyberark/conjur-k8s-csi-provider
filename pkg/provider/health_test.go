package provider

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

type mockResponseWriter struct {
	statusCode   int
	writtenBytes []byte
}

func (w *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.writtenBytes = b
	return len(b), nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

type mockVersionResponder struct {
	called   bool
	response *v1alpha1.VersionResponse
	err      error
}

func (r *mockVersionResponder) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	r.called = true
	return r.response, r.err
}

func TestNewHealthServer(t *testing.T) {
	testCases := []struct {
		description     string
		versionResponse *v1alpha1.VersionResponse
		versionErr      error
		handleFunc      func(http.ResponseWriter, *http.Request)
		assertions      func(*testing.T, *mockVersionResponder, *mockResponseWriter)
	}{
		{
			description:     "provider not serving",
			versionResponse: nil,
			versionErr:      errors.New("some error"),
			assertions: func(t *testing.T, v *mockVersionResponder, w *mockResponseWriter) {
				assert.True(t, v.called)
				assert.Equal(t, 500, w.statusCode)
			},
		},
		{
			description: "happy path",
			versionResponse: &v1alpha1.VersionResponse{
				Version:        "some version",
				RuntimeName:    "some runtime",
				RuntimeVersion: "some runtime version",
			},
			assertions: func(t *testing.T, v *mockVersionResponder, w *mockResponseWriter) {
				assert.True(t, v.called)
				assert.Equal(t, 200, w.statusCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			v := &mockVersionResponder{
				called:   false,
				response: tc.versionResponse,
				err:      tc.versionErr,
			}
			p := &ConjurProviderServer{
				versionFunc: v.Version,
			}
			h := NewHealthServer(p)
			go func() {
				h.Start()
			}()

			req, err := http.NewRequest(
				"GET",
				"http://localhost:8080/healthz",
				strings.NewReader(""),
			)
			assert.Nil(t, err)
			w := &mockResponseWriter{}
			h.server.Handler.ServeHTTP(w, req)

			tc.assertions(t, v, w)

			err = h.Stop()
			assert.Nil(t, err)
		})
	}
}
