NAME=hera
BUILD_IMAGE=$(NAME)-build
COPY_CONTAINER=$(NAME)-copy
RELEASE_NAME=aschaper/$(NAME)
TAG=`cat VERSION`

default: build-all

release: build-all tag push

build-all: build copy build-hera

build-run: build-all run

build: build-image build-binary

build-image:
	docker build -t $(BUILD_IMAGE) -f build.Dockerfile .

build-binary-args:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0

build-binary:
	docker run --rm -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 -it $(BUILD_IMAGE) go build -o /dist/hera

build-hera:
	docker build -t $(NAME):$(TAG) .

test:
	docker run --rm -it $(BUILD_IMAGE) go test -v

copy:
	docker create --name $(COPY_CONTAINER) $(BUILD_IMAGE)
	docker cp $(COPY_CONTAINER):/dist $(shell pwd)
	docker rm $(COPY_CONTAINER)

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd)/.cloudflared:/root/.cloudflared $(NAME):$(TAG)

tunnel:
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=$(NAME) nginx

tag:
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):$(TAG)
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest
	docker push $(RELEASE_NAME):$(TAG)

