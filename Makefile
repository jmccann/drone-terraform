deps:
	GO15VENDOREXPERIMENT=1 go get -u ./...

build:
	GO15VENDOREXPERIMENT=1 GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

clean:
	rm drone-terraform
