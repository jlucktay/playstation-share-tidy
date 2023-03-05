FROM --platform=$BUILDPLATFORM golang:1.20 AS builder
ARG TARGETOS TARGETARCH

# Set some shell options for using pipes and such.
SHELL [ "/bin/bash", "-euo", "pipefail", "-c" ]

# Copy necessary 'go.mod' and 'go.sum' files for separate Go module downloads.
WORKDIR /go/src/go.jlucktay.dev/playstation-share-dedupe
COPY go.* .

# Download Go modules in a separate step before adding the source code, to prevent invalidation of cached Go modules if
# only our source code is changed and not any dependencies.
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
  GOOS=$TARGETOS GOARCH=$TARGETARCH go mod download

# Copy in all of the source code.
COPY . .

# Compile! With the '--mount' flags below, Go's build cache is kept between builds.
# https://github.com/golang/go/issues/27719#issuecomment-514747274
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
  --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
  GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
  -ldflags="-X 'go.jlucktay.dev/version.builtBy=Docker'" -trimpath -v -o /bin/playstation-share-dedupe

FROM gcr.io/distroless/base:nonroot AS deployable
USER 65532

# Bring binary over.
COPY --from=builder /bin/playstation-share-dedupe /bin/playstation-share-dedupe

VOLUME /workdir
WORKDIR /workdir

ENTRYPOINT [ "/bin/playstation-share-dedupe" ]
CMD [ "--help" ]
