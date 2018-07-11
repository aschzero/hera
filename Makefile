NAME=hera
BUILD_IMAGE=$(NAME)-build
RELEASE_NAME=aschzero/$(NAME)
VERSION=`cat VERSION`
PWD=$(shell pwd)

# Build args
GOOS=linux
GOARCH=amd64
CGO_ENABLED=0

default: build

release: version tag push

build: build-binary-image build-binary build-image

build-binary-image:
	docker build -t $(BUILD_IMAGE) -f build.Dockerfile .

build-binary:
	docker run --rm -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -e CGO_ENABLED=$(CGO_ENABLED) -v $(PWD)/dist:/dist -it $(BUILD_IMAGE) go build -o /dist/hera

build-image:
	docker build -t $(NAME) .

test:
	docker run --rm -v $(PWD)/hera:/hera -w /hera $(BUILD_IMAGE) go test -v

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(PWD)/.cloudflared:/root/.cloudflared $(NAME)

tunnel:
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=$(NAME) nginx

version:
	docker run --rm -v $(PWD)/dist:/dist $(BUILD_IMAGE) /dist/$(NAME) -version > VERSION

tag:
	docker tag $(NAME):latest $(RELEASE_NAME):$(VERSION)
	docker tag $(RELEASE_NAME):$(VERSION) $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest
	docker push $(RELEASE_NAME):$(VERSION)

