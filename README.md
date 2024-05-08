# Conjur Provider for Secrets Store CSI Driver

Conjur's integration for the
[Kubernetes Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/),
which injects secrets into Kubernetes environments via
[Container Storage Interface](https://kubernetes-csi.github.io/docs/) volumes.

  * [Certification level](#certification-level)
  * [Requirements](#requirements)
  * [Usage](#usage)
  * [Configuration](#configuration)
    + [Conjur Provider Helm chart](#conjur-provider-helm-chart)
    + [`SecretProviderClass`](#-secretproviderclass-)
  * [Contributing](#contributing)
  * [Community Support](#community-support)
  * [Code Maintainers](#code-maintainers)
  * [License](#license)

<!---<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>--->

Conjur Provider for Secrets Store CSI Driver is part of the CyberArk Conjur
[Open Source Suite](https://cyberark.github.io/conjur/) of tools.

## Certification level

![](https://img.shields.io/badge/Certification%20Level-Trusted-28A745?link=https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md)

This repo is a **Trusted** level project. It is supported by CyberArk and has
been verified to work with Conjur Enterprise. For more detailed information on
our certification levels, see
[our community guidelines](https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md#trusted).

## Requirements

| Dependency               | Minimum Version |
|--------------------------|-----------------|
| Go                       | 1.22.0          |
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
   
    > **Note**  
    > Currently, use of the `token-app-property` variable is not supported.

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
   [documentation](https://docs.cyberark.com/conjur-enterprise/latest/en/Content/Operations/Services/cjr-authn-jwt-lp.htm)
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
       conjur.org/configurationVersion: 0.2.0
       account: myAccount
       applianceUrl: http://myorg.conjur.com
       authnId: authn-jwt/kube
       identity: host/workload-host
       sslCertificate: |
         -----BEGIN CERTIFICATE-----
         MIIDhDCCAmy...njemCrVXIWw==
         -----END CERTIFICATE-----
   ```

   See the [`SecretProviderClass` configuration table](#secretproviderclass) for
   additional customization options.

5. Deploy an application

   Define secrets in the application pod's `conjur.org/secrets` annotation and
   reference the `SecretProviderClass` in the pod's volumes.

```yaml
  ---
  apiVersion: v1
  kind: Pod
  metadata:
    name: app
    namespace: app-namespace
    annotations:
      conjur.org/secrets: |
        - "relative/path/fileA.txt": "db-credentials/url"
        - "relative/path/fileB.txt": "db-credentials/username"
        - "relative/path/fileC.txt": "db-credentials/password"
  spec:
    serviceAccountName: default
    containers:
      - name: app
        image: alpine:latest
        imagePullPolicy: Always
        command: [ "/bin/sh", "-c", "--" ]
        args: [ "while true; do sleep 30; done;" ]
        volumeMounts:
          - name: conjur-csi-provider-volume
            mountPath: /mnt/secrets-store
            readOnly: true
        securityContext:
          allowPrivilegeEscalation: false
    volumes:
      - name: conjur-csi-provider-volume
        csi:
          driver: 'secrets-store.csi.k8s.io'
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
| `daemonSet.name` | Name given to Provider DaemonSet and child Pods | `conjur-k8s-csi-provider` |
| `daemonSet.image.repo` | Conjur Provider Docker image repository | `cyberark/conjur-k8s-csi-provider` |
| `daemonSet.image.tag` | Conjur Provider Docker image tag | `latest` |
| `daemonSet.image.pullPolicy` | Pull Policy for Conjur Provider Docker image | `IfNotPresent` |
| `provider.name` | Name used to reference Conjur Provider instance | `conjur` |
| `provider.healthPort` | Port to expose Conjur Provider health server | `8080` |
| `provider.socketDir` | Directory of socket connections to the Secrets Store CSI Driver | `/var/run/secrets-store-csi-providers` |
| `securityContext` | Security configuration to be applied to Conjur Provider container | <pre>{<br> privileged: false,<br>  allowPrivilegeEscalation: false<br>}</pre> |
| `serviceAccount.create` | Controls whether or not a ServiceAccout is created | `true` |
| `serviceAccount.name` | Name of the ServiceAccount associated with Provider Pods | `conjur-k8s-csi-provider` |
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
| `spec.parameters.conjur.org/configurationVersion` | Conjur CSI Provider configuration version | `0.2.0` |
| `spec.parameters.identity` | Conjur identity used during authentication and authorization | `botApp` |
| `spec.parameters.secrets` | Multiline string describing map of relative filepaths to Conjur variable IDs. NOTE: This parameter is ignored when `conjur.org/configurationVersion` is 0.2.0 or higher. Instead use application pod annotations. | <pre>- "relative/path/fileA.txt": "conjur/path/varA"<br>- "relative/path/fileB.txt": "conjur/path/varB"</pre> |
| `spec.parameters.sslCertificate` | Conjur Appliance certificate | <pre>-----BEGIN CERTIFICATE-----<br>MIIDhDCCAmy...njemCrVXIWw==<br>-----END CERTIFICATE----- |

## Contributing

Please read our [Contributing Guide](CONTRIBUTING.md).

## Community Support

Our primary channel for support is through our CyberArk Commons community
[here](https://discuss.cyberarkcommons.org/c/conjur/5).

## Code Maintainers

CyberArk Conjur Team

## License

Copyright (c) 2023 CyberArk Software Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this software except in compliance with the License. You may obtain a copy of
the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

For the full license text see [LICENSE](./LICENSE).
