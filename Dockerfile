FROM golang:1.10.2-alpine3.7

ADD https://github.com/just-containers/s6-overlay/releases/download/v1.21.4.0/s6-overlay-amd64.tar.gz /tmp/
RUN tar xzf /tmp/s6-overlay-amd64.tar.gz -C /

ADD https://bin.equinox.io/c/VdrWdbjqyF/cloudflared-stable-linux-amd64.tgz /tmp/
RUN tar xzf /tmp/cloudflared-stable-linux-amd64.tgz -C /bin

COPY VERSION /

COPY root /
COPY dist/hera /usr/bin/hera

ENTRYPOINT ["/init"]
