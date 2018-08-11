NAME=hera
BUILDER_IMAGE=$(NAME)-builder
RELEASE_NAME=aschzero/$(NAME)
PWD=$(shell pwd)

default: build

release: tag push

build:
	docker build -t $(NAME) .

test:
	docker build --target builder -t $(BUILDER_IMAGE) .
	docker run --rm $(BUILDER_IMAGE) go test ./...

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(PWD)/.certs:/certs -p 9020:9020 $(NAME)

.PHONY:tunnel
tunnel:
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=$(NAME) nginx

tag:
	docker tag $(NAME):latest $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest

