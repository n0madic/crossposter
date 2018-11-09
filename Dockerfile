FROM golang:alpine

RUN apk add --quiet --no-cache git

ADD . /go/src/github.com/n0madic/crossposter

RUN cd /go/src/github.com/n0madic/crossposter/cmd/crossposter && \
    go get -d -v && \
    go install


FROM alpine

RUN apk add --quiet --no-cache ca-certificates

COPY --from=0 /go/bin/crossposter /usr/bin/

EXPOSE 8000

ENTRYPOINT [ "crossposter" ]
