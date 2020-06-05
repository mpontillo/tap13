build:
	mkdir -p bin/
	go build -o bin/ ./...

clean:
	rm -rf bin/

test: build
	bin/tap13 testdata/*.tap*
