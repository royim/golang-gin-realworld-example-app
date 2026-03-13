---
name: run-qa
description: Run the full quality assurance pipeline (format, vet, lint, test, coverage) and produce a summary report. Use after making code changes or before committing.
argument-hint: "[package path | blank for all]"
allowed-tools: Bash, Read, Grep
---

Run the full QA pipeline and produce a summary report.

**Input**: $ARGUMENTS — optional package path (e.g., `./articles/`). If blank, run on all packages.

---

## Reference

Read `openspec/specs/overview.md` for project tech stack and architecture context if needed.

## Pipeline

Run these checks sequentially. Capture output from each step.

### Step 1: Format Check
```bash
gofmt -l .
```
- PASS: no output (all files formatted)
- FAIL: list of unformatted files

### Step 2: Go Vet
```bash
go vet ./...
```
- PASS: no output
- FAIL: show vet warnings

### Step 3: Lint (if golangci-lint installed)
```bash
which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "SKIPPED: golangci-lint not installed"
```
- PASS: no issues found
- WARN: show warnings with count
- SKIP: tool not installed

### Step 4: Tests
If a specific package is given, test that package. Otherwise test all:
```bash
go test -race -v ./...
```
- Capture: total tests, passed, failed, skipped

### Step 5: Coverage
```bash
go test -coverprofile=/tmp/qa-coverage.out ./... 2>/dev/null
go tool cover -func=/tmp/qa-coverage.out | tail -1
```
- Extract total coverage percentage
- Threshold: 80%

## Report Format

After all steps complete, present this summary:

```
┌─────────────────────────────────────────┐
│            QA Report                    │
├──────────┬──────────────────────────────┤
│ Format   │ PASS / FAIL (N files)       │
│ Vet      │ PASS / FAIL (N issues)      │
│ Lint     │ PASS / WARN / SKIP          │
│ Tests    │ N/N passed                   │
│ Coverage │ XX.X% (>= 80%: PASS/FAIL)  │
├──────────┴──────────────────────────────┤
│ Overall: PASS / FAIL                    │
└─────────────────────────────────────────┘
```

## If Failures Found

For each failure, provide:
1. What failed and why
2. How to fix it (specific command or code change)
3. Offer to auto-fix if possible:
   - Format: `gofmt -w .`
   - Lint: show specific fix suggestions
   - Tests: read the failing test and source code to diagnose

## Guardrails
- Run steps sequentially (later steps depend on earlier ones compiling)
- Do NOT modify any code unless the user asks to fix issues
- Do NOT skip steps even if earlier steps fail (report all issues)
- Coverage threshold is 80% — report as FAIL if below
- Use `-race` flag in tests to detect race conditions
