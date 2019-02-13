# handshake docker testbed

PROJECT := handshake

.PHONY: usage build ssh clean

usage:
	@echo "Targets:"
	@echo "  build: Create a $(PROJECT) docker image"
	@echo "  shell: Run an interactive shell in the container"
	@echo "  clean: Delete all docker images and containers"

build:
	docker build --tag $(PROJECT) .

shell:
	docker run -it $(PROJECT) /bin/bash

clean:
	docker images -aq | xargs -n 1 docker rmi -f
	docker ps -aq | docker rm -f 
