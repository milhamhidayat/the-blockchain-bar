.PHONY: build
build:
	go build ./cmd/tbb/...

.PHONY: migrate
migrate:
	./tbb migrate --datadir=$${HOME}/.tbb

.PHONY: reset-db
reset-db:
	cat /dev/null > $${HOME}/.tbb/database/block.db

.PHONY: api
api:
	./tbb run --datadir=$${HOME}/.tbb
