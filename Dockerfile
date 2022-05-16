FROM golang:alpine AS builder

RUN apk add --quiet --no-cache build-base git

WORKDIR /go/src/github.com/n0madic/crossposter

ADD go.* ./

RUN go mod download

ADD . .

RUN cd cmd/crossposter && \
    go install -ldflags="-linkmode external -extldflags '-static' -s -w"


FROM scratch

EXPOSE 8000

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/* /

ENTRYPOINT [ "/crossposter" ]
