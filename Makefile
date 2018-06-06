NAME=hera
BUILD_IMAGE=$(NAME)-build
COPY_CONTAINER=$(NAME)-copy
RELEASE_NAME=aschaper/$(NAME)
TAG=`cat VERSION`

default: build-all

release: build tag push

build-all: build copy build-image

build-run: build-all run

build:
	docker build -t $(BUILD_IMAGE) -f build.Dockerfile .

build-image:
	docker build -t $(NAME):$(TAG) .

copy:
	docker create --name $(COPY_CONTAINER) $(BUILD_IMAGE)
	docker cp $(COPY_CONTAINER):/dist $(shell pwd)
	docker rm $(COPY_CONTAINER)

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd)/.cloudflared:/root/.cloudflared $(NAME):$(TAG)

tunnel:
	docker run --name=nginx --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=hera nginx

tag:
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):$(TAG)
	docker tag $(NAME):$(TAG) $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest
	docker push $(RELEASE_NAME):$(TAG)

