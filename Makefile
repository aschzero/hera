PWD=$(shell pwd)

IMAGE=audibleblink/hera
BUILDER_IMAGE=$(IMAGE)-builder
CF_URL=https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
CF_BIN=rootfs/bin/cloudflared
TAG ?= latest


help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release: image push ## See image | push

image: $(CF_BIN) ## Buid the Docker image
	docker build -t ${IMAGE}:${TAG} .

push: ## Push to hub.docker.com
	docker push ${IMAGE}:${TAG}

run: ## Run Hera from CWD
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(PWD)/.certs:/certs --network=hera $(IMAGE)

tunnel: ## Create an nginx container that uses Hera. Requires ${HOSTNAME}
	docker run --rm --label hera.hostname=$(HOSTNAME) --label hera.port=80 --network=hera nginx

test: ## Run tests
	docker build --target builder -t $(BUILDER_IMAGE) .
	docker run --rm  $(BUILDER_IMAGE) go test

cfupdate: clean $(CF_BIN) ## Download the latest cloudflared

clean:
	rm ${CF_BIN} hera

$(CF_BIN):
	curl -L -s ${CF_URL} -o ${CF_BIN} && chmod +x ${CF_BIN}

.PHONY: release image test run tunnel push cfupdate clean
