.PHONY: build run test test-v test-race test-e2e test-cover lint lint-new lint-fix fmt vet check qa clean setup-hooks

# ── Build & Run ──

build:
	go build -o bin/server ./...

run:
	go run hello.go

# ── Test ──

test:
	go test ./...

test-v:
	go test -v ./...

test-race:
	go test -race ./...

test-e2e:
	go test -v -count=1 . -run TestE2E

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	@echo "──────────────────────────────"
	@echo "Detail: go tool cover -html=coverage.out"

# ── Lint & Format ──

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# ── Pre-commit Check (new-issues lint + vet + test) ──

check: vet lint-new test
	@echo ""
	@echo "✓ Pre-commit check passed"

lint-new:
	golangci-lint run --new-from-rev=HEAD ./...

# ── QA Pipeline (format → vet → lint → test → coverage) ──

qa: fmt vet lint test-race test-cover
	@echo ""
	@echo "════════════════════════════════"
	@echo "  QA Pipeline Complete"
	@echo "════════════════════════════════"

# ── Utility ──

clean:
	rm -f bin/server coverage.out
	rm -f data/gorm_test.db

tidy:
	go mod tidy

# ── Git Hooks ──

setup-hooks:
	@/bin/cp scripts/pre-commit .git/hooks/pre-commit
	@/bin/chmod +x .git/hooks/pre-commit
	@echo "✓ pre-commit hook installed"
	@/bin/cp scripts/pre-push .git/hooks/pre-push
	@/bin/chmod +x .git/hooks/pre-push
	@echo "✓ pre-push hook installed"
