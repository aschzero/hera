NAME=hera
BUILDIMAGE=$(NAME)-build
INSTANCE=$(NAME)-instance
RELEASENAME=aschaper/$(NAME)
TAG=`cat VERSION`

default: build

dev: build run

build: buildgo copy buildimage

release: build tag push

buildgo:
	docker build -t $(BUILDIMAGE) -f build.Dockerfile .

buildimage:
	docker build -t $(NAME):$(TAG) .

copy:
	docker create --name $(INSTANCE) $(BUILDIMAGE)
	docker cp $(INSTANCE):/dist $(shell pwd)
	docker rm $(INSTANCE)

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd)/.cloudflared:/root/.cloudflared $(NAME):$(TAG)

tag:
	docker tag $(NAME):$(TAG) $(RELEASENAME):$(TAG)
	docker tag $(NAME):$(TAG) $(RELEASENAME):latest

push:
	docker push $(RELEASENAME):latest
	docker push $(RELEASENAME):$(TAG)