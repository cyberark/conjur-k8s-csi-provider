###############
# BUILD STAGE #
###############
FROM golang:1.24-alpine AS builder

# this value changes in ./bin/build
ARG TAG_SUFFIX="dev"
ARG VERSION="unreleased"

# On CyberArk dev laptops, golang module dependencies are downloaded with a
# corporate proxy in the middle. For these connections to succeed we need to
# configure the proxy CA certificate in build containers.
#
# To allow this script to also work on non-CyberArk laptops where the CA
# certificate is not available, we copy the (potentially empty) directory
# and update container certificates based on that, rather than rely on the
# CA file itself.
COPY build_ca_certificate /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /conjur-k8s-csi-provider
COPY . .
RUN go build \
    -ldflags="-X 'github.com/cyberark/conjur-k8s-csi-provider/pkg/provider.ProviderVersion=$VERSION' \
      -X 'github.com/cyberark/conjur-k8s-csi-provider/pkg/provider.TagSuffix=$TAG_SUFFIX'" \
    -o /conjur-csi-provider \
    ./cmd/conjur-k8s-csi-provider/main.go

#############
# RUN STAGE #
#############
FROM alpine:3.19.1 as conjur-k8s-csi-provider
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."
LABEL id="conjur-k8s-csi-provider"

COPY --from=builder /conjur-csi-provider /conjur-csi-provider

ENTRYPOINT [ "/conjur-csi-provider" ]


################
# REDHAT IMAGE #
################
FROM registry.access.redhat.com/ubi9/ubi as conjur-k8s-csi-provider-redhat

ARG VERSION

LABEL org.opencontainers.image.authors="CyberArk Software Ltd."
LABEL id="conjur-k8s-csi-provider"
LABEL name="Conjur Provider for Kubernetes Secrets Store CSI Driver"
LABEL maintainer="CyberArk Software Ltd."
LABEL vendor="CyberArk"
LABEL version="$VERSION"
LABEL release="$VERSION"
LABEL summary="Inject Conjur secrets into Kubernetes environments via Container Storage Interface volumes."
LABEL description="Conjur's integration for the Kubernetes Secrets Store CSI Driver, which injects secrets into \
Kubernetes environments via Container Storage Interface volumes."

RUN yum -y distro-sync

# Add a non-root user with permissions on the default socket dir.
# NOTE: If deploying this image via the helm chart, the csi-provider
# user will require special permissions on the host to access the
# secrets-store-csi-provider socket directory which is volume mounted.
RUN useradd -m csi-provider && \
    mkdir -p /var/run/secrets-store-csi-providers && \
    chown -R csi-provider:0 /var/run/secrets-store-csi-providers

USER csi-provider

COPY LICENSE /licenses/LICENSE
COPY --from=builder /conjur-csi-provider /conjur-csi-provider

ENTRYPOINT [ "/conjur-csi-provider" ]
