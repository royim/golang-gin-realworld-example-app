# Vibe Coding 전환 계획서

> 대상: golang-gin-realworld-example-app
> 작성일: 2026-03-13
> 목적: 기존 Go REST API 프로젝트에 바이브 코딩 워크플로우를 도입하여 AI 협업 개발 환경 구축

---

## 현황 분석

### 프로젝트 상태
- **스택**: Go 1.21+, Gin v1.10, GORM v2, SQLite, JWT v5
- **구조**: 도메인 기반 패키지 (users/, articles/, common/)
- **테스트**: 통합 테스트 존재 (httptest + 실제 SQLite)
- **문서화**: CLAUDE.md 존재, OpenSpec 초기 설정 완료 (미구성)
- **CI/CD**: GitHub Actions 없음, Git hook 없음
- **스킬**: OpenSpec opsx 스킬 4종 존재 (propose, apply, archive, explore)

### 갭 분석
| 영역 | 현재 | 목표 |
|------|------|------|
| API 문서화 | CLAUDE.md만 존재 | OpenSpec 전체 API 스펙 |
| 스킬 | opsx 4종만 | 개발 워크플로우 전체 커버 |
| 테스트 커버리지 | 부분적 | 80%+ (단위 + 통합 + E2E) |
| 린트 | 없음 | golangci-lint 통합 |
| 빌드 자동화 | 없음 | Makefile 전체 타겟 |
| Git hook | 없음 | lefthook (pre-commit + pre-push) |
| CI | 없음 | GitHub Actions (lint + test + build) |

---

## Epic 1: 문서화 — OpenSpec 스펙 문서화 + 스킬 생성

### 1.1 OpenSpec 설정 완성
**목표**: openspec/config.yaml에 프로젝트 컨텍스트와 규칙 추가

- [ ] `config.yaml`에 tech stack, conventions 기입
  ```yaml
  context: |
    Tech stack: Go 1.21+, Gin v1.10, GORM v2, SQLite, JWT v5
    Domain: RealWorld "Conduit" blogging platform
    Architecture: Domain-based packages (users/, articles/, common/)
    Auth: JWT with "Token" scheme (not Bearer)
    Testing: Integration tests with real SQLite DB
    Conventions: 5-file convention per domain package
  rules:
    proposal:
      - Reference RealWorld spec for API contracts
      - Include affected packages in scope
    tasks:
      - Each task should map to a single package change
      - Include test requirements per task
  ```

### 1.2 API 스펙 문서화
**목표**: 전체 API 엔드포인트를 OpenSpec specs/로 문서화

- [ ] **인증 API 스펙** — `specs/auth.md`
  - POST /api/users (회원가입)
  - POST /api/users/login (로그인)
  - GET /api/user (현재 사용자)
  - PUT /api/user (사용자 정보 수정)

- [ ] **프로필 API 스펙** — `specs/profiles.md`
  - GET /api/profiles/:username
  - POST /api/profiles/:username/follow
  - DELETE /api/profiles/:username/follow

- [ ] **아티클 API 스펙** — `specs/articles.md`
  - GET /api/articles (목록, 필터링, 페이지네이션)
  - GET /api/articles/feed (피드)
  - GET /api/articles/:slug (단일)
  - POST /api/articles (생성)
  - PUT /api/articles/:slug (수정)
  - DELETE /api/articles/:slug (삭제)

- [ ] **댓글 API 스펙** — `specs/comments.md`
  - GET /api/articles/:slug/comments
  - POST /api/articles/:slug/comments
  - DELETE /api/articles/:slug/comments/:id

- [ ] **즐겨찾기 API 스펙** — `specs/favorites.md`
  - POST /api/articles/:slug/favorite
  - DELETE /api/articles/:slug/favorite

- [ ] **태그 API 스펙** — `specs/tags.md`
  - GET /api/tags

- [ ] **데이터 모델 스펙** — `specs/models.md`
  - UserModel, ArticleModel, TagModel, CommentModel, FavoriteModel, FollowModel 관계 정의

### 1.3 개발 워크플로우 스킬 생성
**목표**: .github/skills/에 바이브 코딩 전용 스킬 추가

- [ ] **`generate-endpoint`** — 새 API 엔드포인트 생성 스킬
  - 5-file convention에 맞춰 models, routers, serializers, validators, tests 생성
  - 입력: 엔드포인트 설명 → 출력: 패키지 내 전체 파일

- [ ] **`generate-test`** — 테스트 코드 생성 스킬
  - 기존 핸들러에 대한 테스트 코드를 TDD 패턴으로 생성
  - httptest + 실제 SQLite DB 패턴 적용

- [ ] **`run-qa`** — 품질 검사 통합 스킬
  - golangci-lint + go test + coverage 순차 실행
  - 결과를 요약하여 리포트 생성

- [ ] **`generate-plan`** — 마이그레이션/기능 계획 생성 스킬
  - 변경 사항을 분석하고 단계별 계획 문서 생성

- [ ] **`review-code`** — 코드 리뷰 스킬
  - Go 컨벤션, 보안, 성능 관점에서 변경 사항 리뷰

### 1.4 프롬프트 파일 추가
**목표**: .github/prompts/에 워크플로우 프롬프트 추가

- [ ] `generate-endpoint.prompt.md` — 엔드포인트 생성 프롬프트
- [ ] `generate-test.prompt.md` — 테스트 생성 프롬프트
- [ ] `run-qa.prompt.md` — QA 실행 프롬프트
- [ ] `generate-plan.prompt.md` — 계획 생성 프롬프트
- [ ] `review-code.prompt.md` — 코드 리뷰 프롬프트

---

## Epic 2: 테스트 구현 — 단위 테스트, E2E, 린트, Makefile

### 2.1 테스트 커버리지 확대
**프레임워크**: Go 표준 testing + testify assertion

- [ ] **common/ 패키지 단위 테스트 보강**
  - `utils.go`: GenToken, ExtractToken, HeaderTokenMock 테스트
  - `database.go`: DB 초기화, 마이그레이션 테스트
  - `errors.go`: 에러 핸들링 유틸리티 테스트

- [ ] **users/ 패키지 테스트 보강**
  - models.go: CRUD 함수 단위 테스트
  - routers.go: 각 핸들러별 정상/에러 케이스
  - validators.go: 입력 검증 경계값 테스트
  - middlewares.go: AuthMiddleware 동작 테스트

- [ ] **articles/ 패키지 테스트 보강**
  - models.go: CRUD + 배치 함수 단위 테스트
  - routers.go: 각 핸들러별 정상/에러 케이스
  - serializers.go: 응답 직렬화 테스트
  - validators.go: 입력 검증 경계값 테스트

### 2.2 E2E API 통합 테스트
**방식**: httptest + 실제 SQLite DB, 전체 API 플로우 검증

- [ ] **`e2e_test.go`** (루트 또는 별도 패키지)
  - 회원가입 → 로그인 → 프로필 조회 플로우
  - 아티클 생성 → 수정 → 삭제 플로우
  - 댓글 생성 → 조회 → 삭제 플로우
  - 즐겨찾기 추가 → 목록 확인 → 제거 플로우
  - 팔로우 → 피드 확인 → 언팔로우 플로우
  - 인증 실패/권한 없음 에러 플로우

### 2.3 golangci-lint 설정
- [ ] `.golangci.yml` 설정 파일 생성
  ```yaml
  linters:
    enable:
      - errcheck
      - gosimple
      - govet
      - ineffassign
      - staticcheck
      - unused
      - gosec        # 보안 검사
      - bodyclose    # HTTP body close 검사
      - gofmt
      - goimports
  linters-settings:
    gosec:
      excludes:
        - G101  # 하드코딩 크리덴셜 (테스트 코드)
  issues:
    exclude-dirs:
      - data
  ```

### 2.4 Makefile 생성
- [ ] 전체 타겟이 포함된 `Makefile` 작성

| 타겟 | 명령 | 설명 |
|------|------|------|
| `make run` | `go run hello.go` | 서버 실행 |
| `make build` | `go build -o bin/server ./...` | 바이너리 빌드 |
| `make test` | `go test ./...` | 전체 테스트 실행 |
| `make test-v` | `go test -v ./...` | 상세 테스트 실행 |
| `make coverage` | `go test -coverprofile=... ./...` | 커버리지 리포트 |
| `make e2e` | `go test -tags=e2e ./e2e/...` | E2E 테스트 실행 |
| `make lint` | `golangci-lint run` | 린트 실행 |
| `make fmt` | `gofmt -w .` | 코드 포맷팅 |
| `make vet` | `go vet ./...` | 정적 분석 |
| `make clean` | `rm -rf bin/ coverage.out` | 빌드 산출물 정리 |
| `make check` | `make fmt lint vet test` | 전체 품질 검사 |
| `make ci` | `make lint test coverage` | CI 파이프라인 로컬 실행 |

---

## Epic 3: CI/CD 파이프라인 — Git Hook + GitHub Actions

### 3.1 lefthook 설정
- [ ] `go install github.com/evilmartians/lefthook@latest`
- [ ] `lefthook.yml` 생성

```yaml
pre-commit:
  parallel: true
  commands:
    fmt:
      glob: "*.go"
      run: gofmt -l {staged_files} && test -z "$(gofmt -l {staged_files})"
      fail_text: "gofmt 포맷팅이 필요합니다. 'make fmt' 실행 후 다시 커밋하세요."
    lint:
      glob: "*.go"
      run: golangci-lint run --new-from-rev HEAD
      fail_text: "린트 오류가 있습니다. 'make lint'로 확인하세요."
    vet:
      glob: "*.go"
      run: go vet ./...
      fail_text: "go vet 오류가 있습니다."

pre-push:
  commands:
    test:
      run: go test ./...
      fail_text: "테스트 실패. 'make test'로 확인하세요."
    coverage:
      run: |
        go test -coverprofile=/tmp/coverage.out ./... > /dev/null 2>&1
        COVERAGE=$(go tool cover -func=/tmp/coverage.out | grep total | awk '{print $3}' | tr -d '%')
        if [ $(echo "$COVERAGE < 80" | bc) -eq 1 ]; then
          echo "커버리지 ${COVERAGE}%는 최소 기준 80%에 미달합니다."
          exit 1
        fi
      fail_text: "테스트 커버리지가 80% 미만입니다."
```

### 3.2 GitHub Actions CI 워크플로우
- [ ] `.github/workflows/ci.yml` 생성

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -race -coverprofile=coverage.out ./...
      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          echo "Total coverage: ${COVERAGE}%"
          if [ $(echo "$COVERAGE < 80" | bc) -eq 1 ]; then
            echo "::error::Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go build -o bin/server ./...
```

### 3.3 셋업 자동화
- [ ] `Makefile`에 setup 타겟 추가
  ```makefile
  setup:
  	go install github.com/evilmartians/lefthook@latest
  	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  	lefthook install
  ```

---

## Epic 4: 전환 로드맵

### Phase 1: 기반 구축 (Week 1)
| # | 작업 | 산출물 | 완료 기준 |
|---|------|--------|-----------|
| 1.1 | OpenSpec config.yaml 완성 | `openspec/config.yaml` | 프로젝트 컨텍스트 기입 완료 |
| 1.2 | Makefile 생성 | `Makefile` | `make check` 정상 실행 |
| 1.3 | golangci-lint 설정 | `.golangci.yml` | `make lint` 통과 |
| 1.4 | lefthook 설정 | `lefthook.yml` | `lefthook run pre-commit` 정상 |

### Phase 2: 테스트 강화 (Week 2)
| # | 작업 | 산출물 | 완료 기준 |
|---|------|--------|-----------|
| 2.1 | common/ 테스트 보강 | `common/*_test.go` | 커버리지 80%+ |
| 2.2 | users/ 테스트 보강 | `users/*_test.go` | 커버리지 80%+ |
| 2.3 | articles/ 테스트 보강 | `articles/*_test.go` | 커버리지 80%+ |
| 2.4 | E2E 통합 테스트 작성 | `e2e_test.go` | 핵심 플로우 5개 통과 |

### Phase 3: 문서화 + 스킬 (Week 3)
| # | 작업 | 산출물 | 완료 기준 |
|---|------|--------|-----------|
| 3.1 | API 스펙 문서 작성 | `openspec/specs/*.md` | 전체 엔드포인트 커버 |
| 3.2 | 데이터 모델 스펙 작성 | `openspec/specs/models.md` | 전체 모델 관계 기술 |
| 3.3 | 워크플로우 스킬 생성 | `.github/skills/*` | 5종 스킬 정상 동작 |
| 3.4 | 프롬프트 파일 생성 | `.github/prompts/*` | 5종 프롬프트 완성 |

### Phase 4: CI/CD + 검증 (Week 4)
| # | 작업 | 산출물 | 완료 기준 |
|---|------|--------|-----------|
| 4.1 | GitHub Actions 설정 | `.github/workflows/ci.yml` | PR에서 CI 그린 |
| 4.2 | 전체 워크플로우 검증 | - | 스킬로 기능 1개 E2E 개발 |
| 4.3 | CLAUDE.md 업데이트 | `CLAUDE.md` | 새 워크플로우 반영 |
| 4.4 | 최종 점검 | - | 성공 기준 전항목 통과 |

---

## 성공 기준 (품질 지표 기반)

| 지표 | 기준 | 측정 방법 |
|------|------|-----------|
| 테스트 커버리지 | ≥ 80% | `make coverage` |
| 린트 통과 | 0 errors | `make lint` |
| CI 그린 | 전체 통과 | GitHub Actions 상태 |
| 스킬 동작 | 5종 정상 | 각 스킬 실행 검증 |
| API 스펙 완성도 | 100% 엔드포인트 | 스펙 vs 라우터 대조 |
| Git hook 동작 | pre-commit + pre-push | `lefthook run --all` |
| E2E 테스트 | 핵심 플로우 5개 | `make e2e` |

---

## 위험 요소 및 대응

| 위험 | 영향 | 대응 |
|------|------|------|
| cgo 의존성 (SQLite) | CI 환경에서 빌드 실패 가능 | GitHub Actions에 gcc 설치 단계 추가 |
| 기존 테스트와 신규 테스트 충돌 | DB 격리 실패 | TestDBInit/TestDBFree 패턴 일관 적용 |
| golangci-lint 과도한 경고 | 초기 도입 부담 | 단계적 린터 활성화 (Phase 1은 핵심만) |
| lefthook pre-push 느림 | 개발 속도 저하 | `--no-verify` 옵션 안내 (긴급 시만) |

---

## 부록: 디렉토리 구조 (전환 완료 후)

```
golang-gin-realworld-example-app/
├── .github/
│   ├── prompts/           # Claude Code 프롬프트
│   │   ├── opsx-*.prompt.md      (기존)
│   │   ├── generate-endpoint.prompt.md
│   │   ├── generate-test.prompt.md
│   │   ├── run-qa.prompt.md
│   │   ├── generate-plan.prompt.md
│   │   └── review-code.prompt.md
│   ├── skills/            # Claude Code 스킬
│   │   ├── openspec-*/           (기존)
│   │   ├── generate-endpoint/
│   │   ├── generate-test/
│   │   ├── run-qa/
│   │   ├── generate-plan/
│   │   └── review-code/
│   └── workflows/
│       └── ci.yml
├── openspec/
│   ├── config.yaml        (완성)
│   ├── changes/
│   └── specs/
│       ├── auth.md
│       ├── profiles.md
│       ├── articles.md
│       ├── comments.md
│       ├── favorites.md
│       ├── tags.md
│       └── models.md
├── .golangci.yml
├── lefthook.yml
├── Makefile
├── CLAUDE.md              (업데이트)
└── docs/
    └── migration-plan.md  (본 문서)
```
