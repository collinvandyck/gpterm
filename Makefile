.PHONY: install
install:
	go install ./cmd/gpterm

.PHONY: gpterm
gpterm:
	go run ./cmd/gpterm

.PHONY: repl
repl:
	go run ./cmd/gpterm repl

.PHONY: schema
schema:
	@go run ./cmd/gpterm schema 

.PHONY: sqlc
sqlc:
	@go run github.com/kyleconroy/sqlc/cmd/sqlc generate --file db/sqlc.yaml

.PHONY: clean
clean:
	rm -r ./bin/*
