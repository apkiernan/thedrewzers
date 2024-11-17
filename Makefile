all: tpl styles server

build:
	go build -o bin/main ./cmd/main

server:
	air

tpl:
	templ generate

styles:
	npm run watch

