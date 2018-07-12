FROM alpine:3.8

RUN apk add --no-cache \
  ca-certificates \
  curl

RUN curl -L -s https://github.com/just-containers/s6-overlay/releases/download/v1.21.4.0/s6-overlay-amd64.tar.gz \
  | tar xvzf - -C /

RUN curl -L -s https://bin.equinox.io/c/VdrWdbjqyF/cloudflared-stable-linux-amd64.tgz \
  | tar xvzf - -C /bin

RUN apk del --no-cache curl

COPY root /
COPY dist/hera /usr/bin/hera

ENTRYPOINT ["/init"]
