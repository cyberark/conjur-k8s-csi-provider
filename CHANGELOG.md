# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Nothing should go in this section, please add to the latest unreleased version
  (and update the corresponding date), or add a new version.

## [0.1.0] - 2023-01-18

### Fixed
- Fixed an error in Provider termination which prevented the socket used to
  connect to the Secrets Store CSI Driver from being closed and removed.
  [Conjur-Enterprise/conjur-k8s-csi-provider#19](https://github.cyberng.com/Conjur-Enterprise/conjur-k8s-csi-provider/pull/19)

### Added
- Added additional logging to gRPC and HTTP servers.
  [Conjur-Enterprise/conjur-k8s-csi-provider#19](https://github.cyberng.com/Conjur-Enterprise/conjur-k8s-csi-provider/pull/19)
- Helm chart allows for customizing Provider container's `securityContext`.
  [Conjur-Enterprise/conjur-k8s-csi-provider#19](https://github.cyberng.com/Conjur-Enterprise/conjur-k8s-csi-provider/pull/19)
- Provider and Helm chart support customizable socket directory path and health
  server port.
  [Conjur-Enterprise/conjur-k8s-csi-provider#19](https://github.cyberng.com/Conjur-Enterprise/conjur-k8s-csi-provider/pull/19)

### Changed
- Docker image now built from Alpine base image.
  [Conjur-Enterprise/conjur-k8s-csi-provider#19](https://github.cyberng.com/Conjur-Enterprise/conjur-k8s-csi-provider/pull/19)

## [0.0.1] - 2023-12-26

### Added
- Initial release of Conjur Provider for Secrets Store CSI Driver
