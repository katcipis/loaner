ARG GOVERSION

FROM golang:${GOVERSION} as base

ARG VERSION

WORKDIR /build

COPY . .

RUN go build -o loaner -ldflags "-X main.VersionString=${VERSION}" ./cmd/loaner/loaner.go

# Use two stages only to avoid source code on final image
FROM golang:${GOVERSION}

COPY --from=base /build/loaner /app/loaner

ENTRYPOINT ["/app/loaner"]
