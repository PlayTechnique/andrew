build:
	cd cmd/andrew && go build

run:
	./cmd/andrew/andrew

test:
	docker run --rm -v "$${PWD}":/usr/src/andrew -w /usr/src/andrew golang:1.22 go test