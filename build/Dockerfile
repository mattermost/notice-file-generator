# This dockerfile is used to build Mattermost notice-filep-generator
# A multi stage build, with golang used as a builder
# and gcr.io/distroless/static as runner
ARG GO_IMAGE=golang:1.18.3@sha256:b203dc573d81da7b3176264bfa447bd7c10c9347689be40540381838d75eebef
# hadolint ignore=DL3006
FROM ${GO_IMAGE} as builder

#GO_BUILD_PLATFORMS holds the platforms that we will build the docker image against
ARG GO_BUILD_PLATFORMS=linux-amd64

# Setup directories structure and compile
COPY . /src
WORKDIR /src
RUN make go-build

# Shrink final image since we only need the rtcd binary
# and use distroless container image as runner for security
FROM gcr.io/distroless/static@sha256:d6fa9db9548b5772860fecddb11d84f9ebd7e0321c0cb3c02870402680cc315f as runner
COPY --from=builder /src/dist/notice-file-generator-linux-amd64 /opt/notice-file-generator/bin/notice-file-generator

# We should refrain from running as privileged user
# Run as UID for nobody
USER 65534

WORKDIR /opt/notice-file-generator/bin/
ENTRYPOINT ["./notice-file-generator"]