.PHONY: build-tbb
build-tbb:
	go build ./cmd/tbb/...

.PHONY: build-tbb-migrate
build-tbb-migrate:
	go build ./cmd/tbbmigrate/...
