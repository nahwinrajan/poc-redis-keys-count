PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
GOBASE=$(shell pwd)
GOPATH=$(GOBASE)/vendor:$(GOBASE):/home/azer/code/golang # You can remove or change the path after last colon.
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

build-linux:
	GOOS=linux GOARCH=amd64 go build -a -o poc-redis-keys-count main.go

build:
	@CGO_ENABLED=0 go build -a -o poc-redis-keys-count main.go

dockerup:
	docker-compose up --build -d

dockerdown:
	docker-compose down