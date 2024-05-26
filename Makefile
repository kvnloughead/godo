include .env

# ============================================================
# HELPERS
# ============================================================

## help: print this help message
.PHONY: help
help:
	@echo "\nUsage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo "\nFlags: \n"
	@echo "  Command line flags are supported for run/api and run/air.\n  Specify them like this: "
	@echo "\n\t  make FLAGS=\"-x -y\" command"
	@echo "\n  For a list of implemented flags for the ./cmd/api application, \n  run 'make help/web'\n"
	@echo "\nEnvirontmental Variables:\n"
	@echo "  Environmental variables are supported for run/api and run/air.\n  They can be exported to the environment, or stored in a .env file.\n"

## help/web: prints help from ./cmd/api (including flag descriptions)
.PHONY: help/web
help/web:
	@go run ./cmd/api -help

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ============================================================
# DEVELOPMENT
# ============================================================


## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api $(FLAGS)

# Requires global installation: `go install github.com/cosmtrek/ air@latest`  
# and the appropriate environmental variables. 
## run/air: run server using Air for live reloading. 
.PHONY: run/air
run/air:
	air -- $(FLAGS)

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

## db/migrations/new name=$1: generate new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all 'up' migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running all up migrations'
	migrate -path ./migrations -database ${DB_DSN} up
