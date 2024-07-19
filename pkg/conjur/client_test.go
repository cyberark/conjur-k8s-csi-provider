package conjur

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
)

type mockConjurClient struct {
	retrieveBatchSecretsSafeFunc func([]string) (map[string][]byte, error)
}

func (m *mockConjurClient) RetrieveBatchSecretsSafe(ids []string) (map[string][]byte, error) {
	return m.retrieveBatchSecretsSafeFunc(ids)
}

func TestNewClient(t *testing.T) {
	client := NewClient("url", "authn", "account", "identity", "cert")
	config, ok := client.(*Config)
	if !ok {
		t.Fatalf("NewClient did not return a *Config")
	}
	if config.BaseURL != "url" || config.AuthnID != "authn" || config.Account != "account" ||
		config.Identity != "identity" || config.SSLCert != "cert" {
		t.Errorf("NewClient did not set fields correctly")
	}
	if config.clientFactory == nil {
		t.Errorf("clientFactory should not be nil")
	}
}

func TestGetSecrets(t *testing.T) {
	testCases := []struct {
		name           string
		config         Config
		jwt            string
		secretIDs      []string
		mockSecrets    map[string][]byte
		mockError      error
		expectedResult map[string][]byte
		expectedError  string
	}{
		{
			name: "Successful retrieval",
			config: Config{
				BaseURL:  "https://example.com",
				AuthnID:  "authn-jwt/kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:       "jwt-token",
			secretIDs: []string{"secret1", "secret2"},
			mockSecrets: map[string][]byte{
				"default:variable:secret1": []byte("value1"),
				"default:variable:secret2": []byte("value2"),
			},
			expectedResult: map[string][]byte{
				"secret1": []byte("value1"),
				"secret2": []byte("value2"),
			},
		},
		{
			name: "Validation error",
			config: Config{
				BaseURL:  "",
				AuthnID:  "authn-jwt/kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:           "jwt-token",
			secretIDs:     []string{"secret1"},
			expectedError: fmt.Sprintf(logmessages.CKCP030, "Must specify an ApplianceURL"),
		},
		{
			name: "Client factory error",
			config: Config{
				BaseURL:  "https://example.com",
				AuthnID:  "authn-jwt/kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:           "jwt-token",
			secretIDs:     []string{"secret1"},
			mockError:     fmt.Errorf("client factory error"),
			expectedError: fmt.Sprintf(logmessages.CKCP030, "client factory error"),
		},
		{
			name: "Retrieve error",
			config: Config{
				BaseURL:  "https://example.com",
				AuthnID:  "authn-jwt/kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:           "jwt-token",
			secretIDs:     []string{"secret1"},
			mockError:     fmt.Errorf("retrieve error"),
			expectedError: fmt.Sprintf(logmessages.CKCP031, "retrieve error"),
		},
		{
			name: "Empty secret IDs",
			config: Config{
				BaseURL:  "https://example.com",
				AuthnID:  "authn-jwt/kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:            "jwt-token",
			secretIDs:      []string{},
			expectedResult: map[string][]byte{},
		},
		{
			name: "Different AuthnID format",
			config: Config{
				BaseURL:  "https://example.com",
				AuthnID:  "kube",
				Account:  "default",
				Identity: "host/test",
				SSLCert:  "cert",
			},
			jwt:       "jwt-token",
			secretIDs: []string{"secret1"},
			mockSecrets: map[string][]byte{
				"default:variable:secret1": []byte("value1"),
			},
			expectedResult: map[string][]byte{
				"secret1": []byte("value1"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.clientFactory = func(config conjurapi.Config) (ConjurClient, error) {
				if tc.name == "Client factory error" {
					return nil, tc.mockError
				}
				mockClient := &mockConjurClient{
					retrieveBatchSecretsSafeFunc: func(ids []string) (map[string][]byte, error) {
						if tc.name == "Retrieve error" {
							return nil, tc.mockError
						}
						return tc.mockSecrets, nil
					},
				}
				return mockClient, nil
			}

			result, err := tc.config.GetSecrets(tc.jwt, tc.secretIDs)

			if tc.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Expected result %v, got %v", tc.expectedResult, result)
			}
		})
	}
}
