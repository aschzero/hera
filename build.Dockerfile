FROM golang:1.10.2-alpine3.7

RUN apk update && apk add git

RUN go get github.com/op/go-logging && \
go get github.com/docker/docker/client && \
go get github.com/radovskyb/watcher && \
go get github.com/spf13/afero

RUN mkdir /hera && mkdir /dist
ADD ./hera /hera/
WORKDIR /hera
