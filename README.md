# Conjur Provider for Secrets Store CSI Driver

Conjur's integration for the
[Kubernetes Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/),
which injects secrets into Kubernetes environments via
[Container Storage Interface](https://kubernetes-csi.github.io/docs/) volumes.

  * [Requirements](#requirements)
  * [Usage](#usage)
  * [Configuration](#configuration)
    + [Conjur Provider Helm chart](#conjur-provider-helm-chart)
    + [`SecretProviderClass`](#-secretproviderclass-)
  * [Contributing](#contributing)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

## Requirements

| Dependency               | Minimum Version |
|--------------------------|-----------------|
| Go                       | 1.21.0          |
| Kubernetes               | 1.19.0          |
| Secrets Store CSI Driver | 1.3.0           |
| Conjur OSS / Enterprise  | 1.17.3 / 12.5   |

## Usage

1. Create and configure a JWT Authenticator instance in Conjur

   Load the following Conjur policy samples to setup AuthnJWT.

   Each workload in Kubernetes is represented as a Conjur `host`, specified by
   identifying annotations.

   ```yaml
   - !host
     id: workload-host
     annotations:
       authn-jwt/kube/kubernetes.io/namespace: app-namespace
       authn-jwt/kube/kubernetes.io/serviceaccount/name: sa-name
   ```

   The following policy YAML creates an AuthnJWT instance `kube` to authenticate
   workloads in Kubernetes using their ServiceAccount tokens, and permits the
   created `host` to authenticate with the service.

   ```yaml
   - !policy
     id: conjur/authn-jwt/kube
     body:
     - !webservice

     # Uncomment one of following variables depending on the public availability
     # of the Service Account Issuer Discovery service in Kubernetes:
     # If the service is publicly available, uncomment 'jwks-uri'.
     # If the service is not available, uncomment 'public-keys'.
     # - !variable
     #   id: jwks-uri
     - !variable
       id: public-keys

     # Used with 'jwks-uri'.
     # Uncomment ca-cert if the JWKS website cert isn't trusted by conjur
     # - !variable
     #   id: ca-cert

     # Used with 'public-keys'.
     # This variable contains what "iss" in the JWT.
     - !variable
       id: issuer

     # This variable contains what "aud" is the JWT.
     # - !variable
     #   id: audience

     # This variable tells Conjur which claim in the JWT to use to determine the
     # Conjur host identity.
     # - !variable
     #   id: token-app-property # Most likely set to "sub" for Kubernetes
   
     # Used with 'token-app-property'.
     # This variable will hold the Conjur policy path that contains the Conjur
     # host identity found by looking at the claim entered in token-app-property.
     # - !variable
     #   id: identity-path

     - !permit
       role: !host /workload-host
       privilege: [ read, authenticate ]
       resource: !webservice
   ```

   Create variables that contain secret content required by your application,
   and permit the `host` to access them.

   ```yaml
   - !policy
     id: db-credentials
     body:
     - &variables
       - !variable url
       - !variable username
       - !variable password

     - !permit
       role: !host /workload-host
       privileges: [ read, execute ]
       resource: *variables
   ```

   Refer to our
   [documentation](https://docs.cyberark.com/conjur-enterprise/12.5/en/Content/Operations/Services/cjr-authn-jwt-lp.htm)
   for more information on JWT Authentication.

2. Install the Secrets Store CSI Driver Helm chart

   ```shell
   $ helm repo add secrets-store-csi-driver \
       https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
   $ helm install csi-secrets-store \
       secrets-store-csi-driver/secrets-store-csi-driver \
       --wait \
       --namespace kube-system \
       --set 'tokenRequests[0].audience=conjur'
   ```

   Refer to the Secrets Store CSI Driver
   [documentation](https://secrets-store-csi-driver.sigs.k8s.io/introduction)
   for more information and
   [best practices](https://secrets-store-csi-driver.sigs.k8s.io/topics/best-practices)
   for installing the CSI Driver.

3. Install the Conjur Provider Helm chart

   ```shell
   $ helm repo add cyberark \
       https://cyberark.github.io/helm-charts
   $ helm install conjur-csi-provider \
       cyberark/conjur-k8s-csi-provider \
       --wait \
       --namespace kube-system
   ```

   See the [Helm chart configuration table](#conjur-provider-helm-chart) for
   additional customization options.

4. Create a `SecretProviderClass`

   Configuration is passed to the Conjur provider via a
   [`SecretProviderClass`](https://secrets-store-csi-driver.sigs.k8s.io/concepts#secretproviderclass)
   through the `spec.parameters` field.

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
       authnId: authn-jwt/kube
       identity: host/workload-host
       secrets: |
         - "relative/path/fileA.txt": "db-credentials/url"
         - "relative/path/fileB.txt": "db-credentials/username"
         - "relative/path/fileC.txt": "db-credentials/password"
       sslCertificate: |
         -----BEGIN CERTIFICATE-----
         MIIDhDCCAmy...njemCrVXIWw==
         -----END CERTIFICATE-----
   ```

   See the [`SecretProviderClass` configuration table](#secretproviderclass) for
   additional customization options.

5. Deploy an application

   Reference the `SecretProviderClass` in an application pod's volumes.

   ```yaml
   volumes:
   - name: secrets-store-inline
     csi:
       driver: secrets-store.csi.k8s.io
       readOnly: true
       volumeAttributes:
         secretProviderClass: "credentials-from-conjur"
   ```

## Configuration

### Conjur Provider Helm chart

The following table lists the configurable parameters of the Conjur Provider
Helm chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `providerServer.name` | Name given to Provider DaemonSet and child Pods | `conjur-k8s-csi-provider` |
| `providerServer.image.repo` | Conjur Provider Docker image repository | `cyberark/conjur-k8s-csi-provider` |
| `providerServer.image.tag` | Conjur Provider Docker image tag | `latest` |
| `providerServer.image.pullPolicy` | Pull Policy for Conjur Provider Docker image | `IfNotPresent` |
| `labels` | Map of labels applied to Provider DaemonSet and child Pods | `{}` |
| `annotations` | Map of annotations applied to Provider DaemonSet and child Pods | `{}` |

### `SecretProviderClass`

The following table lists the configurable parameters on the Conjur Provider's
`SecretProviderClass` instances.

| Field | Description | Example |
|-------|-------------|---------|
| `spec.parameters.account` | Conjur account used during authentication | `myAccount` |
| `spec.parameters.applianceUrl` | Conjur Appliance URL | `https://myorg.conjur.com` |
| `spec.parameters.authnId` | Type and service ID of desired Conjur authenticator | `authn-jwt/service-id` |
| `spec.parameters.identity` | Conjur identity used during authentication and authorization | `botApp` |
| `spec.parameters.secrets` | Multiline string describing map of relative filepaths to Conjur variable IDs | <pre>- "relative/path/fileA.txt": "conjur/path/varA"<br>- "relative/path/fileB.txt": "conjur/path/varB"</pre> |
| `spec.parameters.sslCertificate` | Conjur Appliance certificate | <pre>-----BEGIN CERTIFICATE-----<br>MIIDhDCCAmy...njemCrVXIWw==<br>-----END CERTIFICATE----- |

## Contributing

Please read our [Contributing Guide](CONTRIBUTING.md).
