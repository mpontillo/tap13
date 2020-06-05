build:
	mkdir -p bin/
	go build -o bin/ ./...

clean:
	rm -rf bin/
