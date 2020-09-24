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

.PHONY: api-babayaga
api-babayaga:
	./tbb run --datadir=$${HOME}/.tbb-babayaga --port=8081

.PHONY: apiv2-andrej
apiv2-andrej:
	./tbb run --datadir=$${HOME}/.andrej_sync --ip=127.0.0.1 --port=8080

.PHONY: apiv2-babayaga
apiv2-babayaga:
	./tbb run --datadir=$${HOME}/.babayaga_sync --ip=127.0.0.1 --port=8081

.PHONY: apiv2-caesar
apiv2-caesar:
	./tbb run --datadir=$${HOME}/.caesar_sync --ip=127.0.0.1 --port=8082
