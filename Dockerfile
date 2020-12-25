ARG GOVERSION

FROM golang:${GOVERSION}

WORKDIR /app

COPY . .

RUN go build -o loaner ./cmd/loaner/loaner.go

ENTRYPOINT ["/app/loaner"]
