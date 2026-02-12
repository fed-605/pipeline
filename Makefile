.PHONY: docker-build docker-run

docker-build:
	docker build -t go-pipeline

docker-run:
	docker run -it --rm go-pipeline

make run:
	go run main.go