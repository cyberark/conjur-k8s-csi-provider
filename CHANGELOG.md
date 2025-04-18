# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Nothing should go in this section, please add to the latest unreleased version
  (and update the corresponding date), or add a new version.

## [0.2.4] - 2025-04-01

## Security
- Bump Golang base images to 1.24 (CNJR-8631)

## [0.2.3] - 2024-12-20

## Security
- Bumped golang.org/x/net to v0.33.0 to address CVE-2024-45338

## [0.2.2] - 2024-12-16

### Added
- Added default resource limits to provider helm chart (CNJR-6443)

## [0.2.1] - 2024-07-23

### Changed
- Use Conjur Go SDK built-in JWT authentication (CNJR-5497)
- Updated log messages when client configuration or authentication fails (CNJR-5497)

## [0.2.0] - 2024-06-05

### Added
- Support retrieving secrets definition from pod annotations (CNJR-4099)
- Added support for configurable log levels using the `LOG_LEVEL` environment
  variable (CNJR-3733)
- Added support for JWT authenticator field `token-app-property`, which makes
  the `identity` configuration attribute optional (CNJR-4607)

### Changed
- Updated Alpine base image to 3.19.1 (CONJSE-1852)
- Updated google.golang.org/grpc to v1.63.2 (CONJSE-1852)
- Updated google.golang.org/protobuf to v1.33.0 (CONJSE-1852)
- Updated golang.org/x/net to v0.24.0 (CONJSE-1852)
- Updated log messages with unique identifiers (CNJR-3733)

## [0.1.2] - 2024-03-22

### Changed
- Updated Go to 1.22 (CONJSE-1842)

## [0.1.1] - 2023-03-12

### Added
- Include a Redhat UBI9 based Docker image. (CNJR-3715)

## [0.1.0] - 2023-01-18

### Added
- Helm chart allows for customizing Provider container's `securityContext`.
- Provider and Helm chart support customizable socket directory path and health
  server port.

### Changed
- Docker image now built from Alpine base image. (CNJR-3722)

## [0.0.2] - 2023-01-22

### Fixed
- Fixed an error in Provider termination which prevented the socket used to
  connect to the Secrets Store CSI Driver from being closed and removed.

### Added
- Added additional logging to gRPC and HTTP servers.

## [0.0.1] - 2023-12-26

### Added
- Initial release of Conjur Provider for Secrets Store CSI Driver

[Unreleased]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.2.4...HEAD
[0.2.4]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/cyberark/conjur-k8s-csi-provider/compare/v0.0.2...v0.1.1
[0.0.2]: https://github.com/cyberark/conjur-k8s-csi-provider/releases/tag/v0.0.2
