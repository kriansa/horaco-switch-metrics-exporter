.PHONY: build clean

dist:
	mkdir dist

dist/switch-exporter: dist
	CGO_ENABLED=0 go build -o dist/switch-exporter ./cmd

build: dist/switch-exporter

clean:
	go clean
	rm -rf dist
