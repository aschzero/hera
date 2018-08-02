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
	docker run --rm -v $(PWD)/hera:/hera -w /hera $(BUILDER_IMAGE) go test -v

run:
	docker run --rm --name=$(NAME) --network=$(NAME) -v /var/run/docker.sock:/var/run/docker.sock -v $(PWD)/certs:/certs $(NAME)

tunnel:
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=$(NAME) nginx

tag:
	docker tag $(NAME):latest $(RELEASE_NAME):$(VERSION)
	docker tag $(RELEASE_NAME):$(VERSION) $(RELEASE_NAME):latest

push:
	docker push $(RELEASE_NAME):latest
	docker push $(RELEASE_NAME):$(VERSION)

