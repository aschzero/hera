## Builder image
FROM golang:1.17-alpine AS builder
WORKDIR /src
COPY go.mod go.mod .
RUN go mod download
COPY . .
RUN  go build -o /dist/hera


## Final image
FROM oznu/s6-alpine
COPY rootfs /
COPY --from=builder /dist/hera /bin/
ENTRYPOINT ["/init"]
