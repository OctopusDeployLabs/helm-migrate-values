.PHONY: build
build:
	go build -o bin/migrate-values .

.PHONY: run-help
run: build
	./bin/migrate-values -h

.PHONY: clean
clean:
	rm -f bin/*