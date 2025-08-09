FROM golang:1.24.6-alpine AS download_deps

WORKDIR /src

COPY src/go.mod /src
COPY src/go.sum /src

RUN go mod download


FROM golang:1.24.6-alpine

WORKDIR /src

COPY --from=download_deps /go/pkg/mod /go/pkg/mod

COPY src /src

RUN go build

ENTRYPOINT ["./bungalow"]