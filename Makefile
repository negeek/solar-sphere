MODULES := solar-spectrum solar-auth solar-sentinel solar-galaxy

.DEFAULT_GOAL := help
.PHONY: help all build test test-integration vet fmt fmt-check tidy \
        keygen migrate-auth migrate-sentinel \
        run-auth run-sentinel run-galaxy \
        docker-up docker-down docker-sample clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: fmt-check vet test build ## Format-check, vet, test, and build every module (what CI should run)

build: ## go build ./... in every module
	@for m in $(MODULES); do \
		echo "==> build $$m"; \
		(cd $$m && go build ./...) || exit 1; \
	done

test: ## go test ./... in every module (DB-backed tests skip unless TEST_DATABASE_URL is set)
	@for m in $(MODULES); do \
		echo "==> test $$m"; \
		(cd $$m && go test ./...) || exit 1; \
	done

test-integration: ## Same as test, but against a real MongoDB (TEST_DATABASE_URL, default mongodb://localhost:27017)
	@url=$${TEST_DATABASE_URL:-mongodb://localhost:27017}; \
	for m in $(MODULES); do \
		echo "==> test (integration) $$m"; \
		(cd $$m && TEST_DATABASE_URL=$$url go test ./...) || exit 1; \
	done

vet: ## go vet ./... in every module
	@for m in $(MODULES); do \
		echo "==> vet $$m"; \
		(cd $$m && go vet ./...) || exit 1; \
	done

fmt: ## gofmt every module in place
	@gofmt -l -w $(MODULES)

fmt-check: ## Fail if anything isn't gofmt-formatted (no writes; what CI should run)
	@unformatted=$$(gofmt -l $(MODULES)); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

tidy: ## go mod tidy in every module
	@for m in $(MODULES); do \
		echo "==> tidy $$m"; \
		(cd $$m && go mod tidy) || exit 1; \
	done

keygen: ## Print a fresh SIGNING_KEY/VERIFICATION_KEY pair
	go run ./solar-spectrum/cmd/keygen

migrate-auth: ## Apply solar-auth's database migrations
	cd solar-auth && go run ./db/v1

migrate-sentinel: ## Apply solar-sentinel's database migrations
	cd solar-sentinel && go run ./db/v1

run-auth: ## Run solar-auth locally (needs solar-auth/.env, APP_ENV=dev)
	cd solar-auth && go run .

run-sentinel: ## Run solar-sentinel locally (needs solar-sentinel/.env, APP_ENV=dev)
	cd solar-sentinel && go run .

run-galaxy: ## Run solar-galaxy locally (needs solar-galaxy/.env, APP_ENV=dev)
	cd solar-galaxy && go run .

docker-up: ## Build images from source and start the full stack
	docker compose up --build

docker-down: ## Stop the stack started by docker-up
	docker compose down

docker-sample: ## Start the full stack using the pre-built published images
	docker compose -f docker-compose.sample.yml up

clean: ## Remove local build artifacts
	@for m in $(MODULES); do rm -rf $$m/bin; done
