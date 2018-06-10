NAME=hera
BUILD_IMAGE=$(NAME)-build
COPY_CONTAINER=$(NAME)-copy
RELEASE_NAME=aschaper/$(NAME)
TAG=`cat VERSION`
PWD=$(shell pwd)

# Build args
GOOS=linux
GOARCH=amd64
CGO_ENABLED=0

default: build

release: build tag push

build: build-binary-image build-binary build-image

test: build-binary-image run-test

build-binary-image:
	docker build -t $(BUILD_IMAGE) -f build.Dockerfile .

build-binary:
	docker run --rm -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -e CGO_ENABLED=$(CGO_ENABLED) -v $(PWD)/dist:/dist -it $(BUILD_IMAGE) go build -o /dist/hera

build-image:
	docker build -t $(NAME):$(TAG) .

run-test:
	docker run --rm -it $(BUILD_IMAGE) go test -v

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(PWD)/.cloudflared:/root/.cloudflared $(NAME):$(TAG)

tunnel:
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=$(NAME) nginx

tag:
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):$(TAG)
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest
	docker push $(RELEASE_NAME):$(TAG)

