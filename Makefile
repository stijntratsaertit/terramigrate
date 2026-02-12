.PHONY: build test lint clean

build:
	go build -o terramigrate .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f terramigrate
	rm -rf migrations/
