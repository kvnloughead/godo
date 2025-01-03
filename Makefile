ifdef ENV
	ifeq ($(ENV), production)
		include .env.production
		PGHOST=34.134.133.227
	else
		include .env.local
		PGHOST=localhost
	endif
else
	include .env.local
	PGHOST=localhost
endif

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
	@echo "\nEnvironmental Variables:\n"
	@echo "  Environmental variables are supported for run/api and run/air.\n  They can be exported to the environment, or stored in a .env file.\n"

## help/web: prints help from ./cmd/api (including flag descriptions)
.PHONY: help/web
help/web:
	@go run ./cmd/api -help

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ============================================================
# API DEVELOPMENT
# ============================================================

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${DB_DSN} $(FLAGS)

# Requires global installation: `go install github.com/cosmtrek/ air@latest`  
# and the appropriate environmental variables. 
## run/air: run server using Air for live reloading. 
.PHONY: run/air
run/air:
	air -- $(FLAGS)

# ============================================================
# DATABASE SETUP (once per environment)
# ============================================================

## db/setup: create database and user with appropriate permissions
.PHONY: db/setup
db/setup: confirm
	@echo 'Creating database and user...'
	@echo "Running setup script..."
	PGHOST=$(PGHOST) psql -U postgres \
		-v db_name=$(DB_NAME) \
		-v db_user=$(DB_USER) \
		-v db_password='$(DB_PASSWORD)' \
		-v ON_ERROR_STOP=1 \
		-f db/init/setup.sql
	@echo 'Database setup complete.'

# ============================================================
# DATABASE MANAGEMENT
# ============================================================

## db/drop: drop database and user (destructive!)
.PHONY: db/drop
db/drop: confirm
	@echo 'Dropping database and user...'
	@sudo cp db/scripts/drop.sql /tmp/db_drop.sql
	@echo "Setting permissions..."
	@sudo chown postgres:postgres /tmp/db_drop.sql
	@echo "Running drop script..."
	@cd /tmp && sudo -u postgres psql \
		-v db_name=$(DB_NAME) \
		-v db_user=$(DB_USER) \
		-v ON_ERROR_STOP=1 \
		-f /tmp/db_drop.sql
	@echo "Cleaning up..."
	@sudo rm /tmp/db_drop.sql
	@echo 'Database and user dropped.'

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	@psql ${DB_DSN}

## db/backup: backup the database
.PHONY: db/backup
db/backup:
	@echo 'Backing up database...'
	@mkdir -p db/backup
	@pg_dump ${DB_DSN} > db/backup/$(shell date +%Y-%m-%d-%H-%M-%S).sql
	@echo 'Database backup complete.'

## db/migrations/new name=$1: generate new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./db/migrations ${name}

## db/migrations/up: apply all 'up' migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running all up migrations'
	migrate -path ./db/migrations -database ${DB_DSN} up

## db/migrations/up/new: apply only new 'up' migrations
.PHONY: db/migrations/up/new
db/migrations/up/new: confirm
	@echo 'Running new up migrations'
	migrate -path ./db/migrations -database ${DB_DSN} up 1

## db/migrations/down: apply all 'down' migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running all down migrations'
	migrate -path ./db/migrations -database ${DB_DSN} down

## db/migrations/goto version=$1: goto specific migration version
.PHONY: db/migrations/goto
db/migrations/goto: confirm
	@echo 'Migrating to version ${version}'
	migrate -path ./db/migrations -database ${DB_DSN} goto ${version}

## db/migrations/clean version=$1: Intended to clean dirty DB.  Specify the dirty version number, and it will be decremented and passed as an argument to force
.PHONY: db/migrations/clean
db/migrations/clean: confirm
	@echo 'Cleaning DB (dirty version ${version})'
	@decremented_version=$$((${version}-1)); \
	echo "Using version $${decremented_version}"; \
	migrate -path ./db/migrations -database ${DB_DSN} force $${decremented_version}

# ============================================================
# BUILD MANAGEMENT
# ============================================================

## build: build for current OS/architecture in ./build/release
.PHONY: build
build:
	@echo 'Building for current OS/architecture...'
	@mkdir -p build/release
	@go build -o build/release/godo ./cmd/api

## build/linux: build for Linux amd64 (for GCP deployment)
.PHONY: build/linux
build/linux:
	@echo 'Building for Linux amd64...'
	@mkdir -p build/release
	@GOOS=linux GOARCH=amd64 go build -v -o build/release/godo-linux-amd64 ./cmd/api
	@echo 'Build complete: build/release/godo-linux-amd64'

## build/clean: remove build directory
.PHONY: build/clean
build/clean: confirm
	@echo 'Removing build directory...'
	@rm -rf build/

# ============================================================
# DEPLOYMENT MANAGEMENT
# ============================================================

# Deployment targets
.PHONY: deploy/ssh deploy/copy deploy/gcp

## deploy/gcp: stop deployed service, deploy binary, start service again
.PHONY: deploy/gcp
deploy/gcp: build/linux
	@echo 'Deploying to GCP...'
	@ssh ${GCP_USER}@${GCP_HOST} "sudo systemctl stop godo"
	@scp build/release/godo-linux-amd64 ${GCP_USER}@${GCP_HOST}:/opt/godo/
	@ssh ${GCP_USER}@${GCP_HOST} "sudo systemctl start godo"
	@echo 'Deployment complete.'

## deploy/ssh: SSH into the GCP instance
deploy/ssh:
	ssh ${GCP_USER}@${GCP_HOST}

## deploy/copy: Copy file to GCP instance. Usage: make deploy/copy FILE=path/to/file
deploy/copy:
ifndef FILE
	$(error FILE is not set. Usage: make deploy/copy FILE=path/to/file)
endif
	scp ${FILE} ${GCP_USER}@${GCP_HOST}:/opt/godo/

## deploy/psql: connect to the production database using psql
.PHONY: deploy/psql
deploy/psql:
	@echo 'Connecting to production database...'
	@@ENV=production make db/psql

# ============================================================
# CLI INSTALLATION
# ============================================================

## cli/build: builds the CLI application into a binary called 'gd'
.PHONY: cli/build
cli/build:
	@echo 'Building CLI application...'
	@go build -o godo cmd/cli/main.go

## cli/install: installs the CLI application
.PHONY: cli/install
cli/install: cli/build
	@echo 'Installing CLI application...'
	@mkdir -p ${HOME}/.config/godo
	@if [ ! -f ${HOME}/.config/godo/settings.json ]; then \
		echo '{"api_base_url": "http://godo.kevinloughead.com/v1"}' > ${HOME}/.config/godo/settings.json; \
		echo 'Created default config at ~/.config/godo/settings.json'; \
	fi
	@sudo mv godo /usr/local/bin
	@echo 'CLI installation complete'

## cli/setup-local: creates local settings file for development and adds alias to shell config
.PHONY: cli/setup-local
cli/setup-local:
	@echo 'Setting up local development configuration...'
	@mkdir -p ${HOME}/.config/godo
	@if [ ! -f ${HOME}/.config/godo/settings.local.json ]; then \
		echo '{"api_base_url": "http://localhost:4000/v1"}' > ${HOME}/.config/godo/settings.local.json; \
		echo 'Created local config at ~/.config/godo/settings.local.json'; \
	else \
		echo 'Local config already exists at ~/.config/godo/settings.local.json'; \
	fi
	@if ! grep -q "alias gododev=" "${HOME}/.bashrc"; then \
		echo '\n# alias for running godo from the project root with a local config file' >> ${HOME}/.bashrc; \
		echo 'alias gododev='"'"'go run ./cmd/cli/main.go -c=${HOME}/.config/godo/settings.local.json'"'" >> ${HOME}/.bashrc; \
		echo 'Added gododev alias to ~/.bashrc'; \
	else \
		echo 'gododev alias already exists in ~/.bashrc'; \
	fi
	@echo 'Please run: source ~/.bashrc'

.PHONY: cli/logs
## cli/logs opens log files in your preferred editor
cli/logs:
	${EDITOR} ${HOME}/.config/godo/logs/app.log

.PHONY: cli/config
## cli/config opens the user's configuration directory in their preferred editor
cli/config:
	${EDITOR} ${HOME}/.config/godo

.PHONY: token/delete
## token/delete deletes the token file. Specify DEV=true to delete the development token file.
token/delete:
	@if [ "${DEV}" = "true" ]; then \
		rm -f ${HOME}/.config/godo/.token.dev; \
	else \
		rm -f ${HOME}/.config/godo/.token; \
	fi
