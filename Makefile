IMAGE_NAME = ovgucsworkshops/git-userland
TAG = v0.3.2

build:
	docker buildx build -t ${IMAGE_NAME}:latest -t ${IMAGE_NAME}:${TAG} . --push
