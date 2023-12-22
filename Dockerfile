###############
# BUILD STAGE #
###############
FROM golang:1.21-alpine AS builder

# On CyberArk dev laptops, golang module dependencies are downloaded with a
# corporate proxy in the middle. For these connections to succeed we need to
# configure the proxy CA certificate in build containers.
#
# To allow this script to also work on non-CyberArk laptops where the CA
# certificate is not available, we copy the (potentially empty) directory
# and update container certificates based on that, rather than rely on the
# CA file itself.
ADD build_ca_certificate /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /conjur-k8s-csi-provider
ADD . .
RUN go build -o /conjur-csi-provider ./cmd/conjur-k8s-csi-provider/main.go

#############
# RUN STAGE #
#############
FROM scratch
LABEL org.opencontainers.image.authors="CyberArk Software Ltd."
LABEL id="conjur-k8s-csi-provider"

COPY --from=builder /conjur-csi-provider /conjur-csi-provider

ENTRYPOINT [ "/conjur-csi-provider" ]


