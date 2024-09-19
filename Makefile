HELM_PLUGINS := $(shell helm env HELM_PLUGINS)

.PHONY: install
install: build
	mkdir -p $(HELM_PLUGINS)/helm-migrate-values/bin
	cp bin/migrate-values $(HELM_PLUGINS)/helm-migrate-values/bin
	cp plugin.yaml $(HELM_PLUGINS)/helm-migrate-values/

.PHONY: build
build:
	go build -o bin/migrate-values ./cmd/helm-migrate-values

.PHONY: run-help
run: build
	./bin/migrate-values -h

.PHONY: clean
clean:
	rm -f bin/*