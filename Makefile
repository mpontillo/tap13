.PHONY: build
build:
	mkdir -p bin/
	go build -o bin/ ./...

.PHONY: format
format:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: manual-test
manual-test: build
	bin/tap13 testdata/*.tap*

.PHONY: unit-test
unit-test: build
	go test -v ./...

.PHONY: performance-test
performance-test:
	go test -v -bench=. ./...

# Note: The unit-test target is a subset of the performance-test target.
.PHONY: test
test: manual-test performance-test
