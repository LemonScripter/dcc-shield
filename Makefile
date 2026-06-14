all: build

deps:
	go mod download

build: deps
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dcc-shield main.go

clean:
	rm -f dcc-shield
