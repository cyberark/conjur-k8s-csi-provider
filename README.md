# Conjur Provider for Secrets Store CSI Driver

Conjur's integration for the
[Kubernetes Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/),
which injects secrets into Kubernetes environments via
[Container Storage Interface](https://kubernetes-csi.github.io/docs/) volumes.

## Requirements

| Dependency               | Minimum Version |
|--------------------------|-----------------|
| Go                       | 1.21.0          |
| Kubernetes               | 1.19.0          |
| Secrets Store CSI Driver | 1.3.0           |
| Conjur OSS / Enterprise  | 1.17.3 / 12.5   |

## Configuration

Configuration is passed to the Conjur provider via a
[`SecretProviderClass`](https://secrets-store-csi-driver.sigs.k8s.io/concepts#secretproviderclass)
through the `spec.parameters` field. All of the fields described below are
required:

| Field | Description | Example |
|-------|-------------|---------|
| `spec.parameters.account` | Conjur account used during authentication | `myAccount` |
| `spec.parameters.applianceUrl` | Conjur Appliance URL | `https://myorg.conjur.com` |
| `spec.parameters.authnId` | Type and service ID of desired Conjur authenticator | `authn-jwt/service-id` |
| `spec.parameters.identity` | Conjur identity used during authentication and authorization | `botApp` |
| `spec.parameters.secrets` | Multiline string describing map of relative filepaths to Conjur variable IDs | <pre>- "relative/path/fileA.txt": "conjur/path/varA"<br>- "relative/path/fileB.txt": "conjur/path/varB"</pre> |

Applying the example values listed in the table above yields the following
`SecretProviderClass` manifest:

```yaml
---
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: credentials-from-conjur
spec:
  provider: conjur
  parameters:
    account: myAccount
    applianceUrl: http://myorg.conjur.com
    authnId: authn-jwt/service-id
    identity: botApp
    secrets: |
      - "relative/path/fileA.txt": "conjur/path/varA"
      - "relative/path/fileB.txt": "conjur/path/varB"
```

## Usage

Reference the `SecretProviderClass` in an application pod's volumes:

```yaml
volumes:
- name: secrets-store-inline
  csi:
    driver: secrets-store.csi.k8s.io
    readOnly: true
    volumeAttributes:
      secretProviderClass: "credentials-from-conjur"

```

## Contributing

Please read our [Contributing Guide](CONTRIBUTING.md).
