---
name: review-change
description: Review code changes against project conventions, security, performance, and RealWorld spec compliance. Use after modifying code or before committing. Analyzes git diff or specified files.
argument-hint: "[file path | blank for git diff]"
allowed-tools: Read, Grep, Glob, Bash
---

Review code changes against project conventions and best practices.

**Input**: $ARGUMENTS — file path to review, or blank to review current `git diff`.

---

## Reference Specs

Read these specs to understand project standards:

1. `openspec/specs/overview.md` — architecture, error format, status codes
2. `openspec/specs/models.md` — model patterns, relationships, DB functions
3. `openspec/specs/middleware.md` — auth middleware modes, token scheme

Also read the relevant domain spec for endpoints being changed.

## Steps

1. **Get the changes**:
   - If file path given: read that file
   - If blank: run `git diff` (staged + unstaged)
   - If no diff: run `git diff HEAD~1` (last commit)

2. **Analyze each changed file** against these checklists:

### Convention Compliance
- [ ] 5-file convention respected (changes in routers.go have corresponding validator/serializer)
- [ ] Route registration follows project pattern (anonymous vs authenticated groups)
- [ ] Trailing slash duplicate routes included for POST/PUT endpoints
- [ ] Error responses use `common.NewError()` or `common.NewValidatorError()`
- [ ] Uses `common.Bind()` not Gin's default binding

### Auth & Security
- [ ] Protected endpoints use `AuthMiddleware(true)` group
- [ ] Ownership checks for update/delete (compare AuthorID)
- [ ] No hardcoded secrets or credentials
- [ ] User input validated before DB operations
- [ ] SQL injection safe (using GORM parameterized queries)

### Performance
- [ ] List endpoints use `Preload()` for eager loading
- [ ] Batch operations used for N+1 scenarios (`BatchGetFavoriteCounts`, `BatchGetFavoriteStatus`)
- [ ] No unbounded queries (limit/offset for list endpoints)

### RealWorld Spec Compliance
- [ ] Response JSON structure matches spec (nested under `"user"`, `"article"`, `"profile"`, etc.)
- [ ] Status codes match spec (201 for create, 200 for success, 422 for DB errors)
- [ ] Timestamps in RFC3339Nano format
- [ ] Tag list sorted in response
- [ ] Auth header format: `Token <jwt>` (not Bearer)

### Test Coverage
- [ ] New handlers have corresponding test functions
- [ ] Tests cover success + at least one error case
- [ ] Tests use `common.TestDBInit/TestDBFree` pattern
- [ ] Tests use real DB (not mocked)

3. **Produce review** with severity ratings:

   **CRITICAL** — Must fix before merge:
   - Security vulnerabilities (injection, auth bypass)
   - Data loss risks
   - Broken API contract

   **WARNING** — Should fix:
   - Missing error handling
   - Convention violations
   - Missing tests for new handlers
   - N+1 query patterns

   **SUGGESTION** — Nice to have:
   - Code simplification opportunities
   - Better naming
   - Additional edge case tests

## Output Format

```
## Review: <file or scope>

### CRITICAL (N)
- **[file:line]** Description of issue
  Fix: specific suggestion

### WARNING (N)
- **[file:line]** Description of issue
  Fix: specific suggestion

### SUGGESTION (N)
- **[file:line]** Description

### Summary
- Files reviewed: N
- Issues: N critical, N warnings, N suggestions
- Verdict: APPROVE / REQUEST CHANGES
```

## Guardrails
- This is READ-ONLY — do NOT modify any files
- Review against actual project patterns, not generic Go best practices
- Always check auth requirements by looking at route group in hello.go
- Compare response structures against openspec/specs/ domain files
- If no issues found, explicitly state the code looks good
