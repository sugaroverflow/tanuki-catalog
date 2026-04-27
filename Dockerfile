# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
ARG TARGETOS TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /out/tanuki ./cmd/tanuki

FROM alpine:3.22
RUN adduser -D -H tanuki
COPY --from=build /out/tanuki /usr/local/bin/tanuki
COPY catalog.json /home/tanuki/catalog.json
WORKDIR /home/tanuki
USER tanuki
ENTRYPOINT ["tanuki"]
CMD ["list"]
