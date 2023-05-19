FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o shelly-collector github.com/topisenpai/shelly-collector

FROM alpine

COPY --from=build /build/shelly-collector /bin/shelly-collector

EXPOSE 80

ENTRYPOINT ["/bin/shelly-collector"]

CMD ["-config", "/var/lib/shelly-collector/config.yml"]