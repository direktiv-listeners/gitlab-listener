IMAGE := "localhost:5000/gitlab"

.PHONY: cross-prepare
cross-prepare:
	docker buildx create --use      
	docker run --privileged --rm docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

.PHONY: cross-build
cross-build:
	docker buildx build --platform=linux/arm64,linux/amd64 -f Dockerfile --push -t ${IMAGE} .

.PHONY: test
test:
	go test -v  cmd/*.go

.PHONY: push
push:
	docker build -t ${IMAGE} . && docker push ${IMAGE}


