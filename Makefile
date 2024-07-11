.PHONY: build
build:
	go build -o bin/migrate-values ./cmd/helm-migrate-values

.PHONY: run-help
run: build
	./bin/migrate-values -h

.PHONY: clean
clean:
	rm -f bin/*