ARG GOVERSION

FROM golang:${GOVERSION}

ARG VERSION

WORKDIR /app

COPY . .

RUN go build -o loaner -ldflags "-X main.VersionString=${VERSION}" ./cmd/loaner/loaner.go

ENTRYPOINT ["/app/loaner"]
