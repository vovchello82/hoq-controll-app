.DEFAULT_GOAL := build

clean:
	go clean
	rm -rf main
build:
	go clean
	go build main.go
docker:
	docker build -t vzlobins/hoa-control-app:latest .