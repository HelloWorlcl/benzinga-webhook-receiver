.PHONY: test docker-build docker-run

APP_NAME=webhook-receiver
DOCKER_IMAGE=webhook-receiver:latest

test:
	@go test -v -race ./...

fmt:
	@go fmt ./...

lint:
	@docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.6.2 golangci-lint run

docker-build:
	@docker build -t $(DOCKER_IMAGE) .

docker-run:
	@docker run --rm -p 8080:8080 \
		--env-file .env \
		$(DOCKER_IMAGE)
