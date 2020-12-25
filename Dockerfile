ARG VERSION
ARG GOVERSION

FROM golang:${GOVERSION}

WORKDIR /app

COPY . .

ENV VERSION=${VERSION}
RUN echo $GOVERSION
RUN echo $VERSION
RUN go build -o loaner -ldflags "-X main.VersionString=${VERSION}" ./cmd/loaner/loaner.go

ENTRYPOINT ["/app/loaner"]
