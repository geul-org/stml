✅ 완료

# Phase 6: Generator Target 추상화

## 목표

generator를 `Target` 인터페이스 기반으로 재구조화하여 React 코드를 `ReactTarget` 구현체로 격리하고, 새 프레임워크 추가 시 해당 Target만 구현하면 되는 구조를 만든다.

기존 외부 API(`Generate`, `GeneratePage`, `GenerateOptions`, `GenerateResult`)는 시그니처 불변. 기존 테스트 39개 전부 통과해야 한다.

## 근거

수정지시서002. parser IR은 이미 프레임워크 중립이지만 generator가 React를 직접 string-build하고 있다.

---

## Step 1: Target 인터페이스 정의

**새 파일:** `generator/target.go`

```go
type Target interface {
    GeneratePage(page parser.PageSpec, specsDir string, opts GenerateOptions) string
    FileExtension() string
    Dependencies(pages []parser.PageSpec) map[string]string
}
```

`DefaultTarget() Target` → `&ReactTarget{}` 반환.

---

## Step 2: React 코드 이동

파일 이름 변경 및 이동:

| 변경 전 | 변경 후 | 내용 |
|---|---|---|
| `imports.go` | `react_imports.go` | importSet, collectImports, renderImports |
| `templates.go` | `react_templates.go` | renderUseQuery, renderUseMutation, renderChildNodes 등 |
| (generator.go 일부) | `react_target.go` | ReactTarget struct + GeneratePage/FileExtension/Dependencies 메서드 |

**이동 대상 함수 (generator.go → react_target.go):**
- `GeneratePage` 본문 → `(*ReactTarget).GeneratePage()`
- `renderFetchHooks` — React useState/useQuery 생성
- `collectAllActions`, `deduplicateActions` — 공통 유틸이지만 현재 GeneratePage에서만 사용. generator.go에 유지.
- `collectFetchOps`, `collectFetchParamBinds`, `collectAllParams` — generator.go에 유지 (공통 유틸).
- `findRootElement` — generator.go에 유지.

**이동 대상 함수 (imports.go → react_imports.go):**
- `importSet` struct — React 전용
- `collectImports`, `collectFetchImports`, `collectActionImports` — React hook 분석
- `renderImports` — React import 렌더링

**templates.go → react_templates.go:**
- 파일 이름 변경만. 내용 변경 없음.

---

## Step 3: generator.go 리팩토링

generator.go에 남는 것:
- `GenerateOptions`, `GenerateResult`, `DefaultOptions()` — 공통 타입
- `Generate()` — `GenerateWith(DefaultTarget(), ...)` 위임
- `GeneratePage()` — `DefaultTarget().GeneratePage(...)` 위임
- `GenerateWith()` — Target 기반 진입점 (NEW)
- `DefaultTarget()` — `&ReactTarget{}` 팩토리 (NEW)
- 공통 유틸: `toComponentName`, `toUpperFirst`, `toLowerFirst`, `clsAttr`, `collectAllActions`, `deduplicateActions`, `collectAllParams`, `collectFetchOps`, `collectFetchParamBinds`, `findRootElement`

```go
func Generate(pages []parser.PageSpec, specsDir, outDir string, opts ...GenerateOptions) (*GenerateResult, error) {
    return GenerateWith(DefaultTarget(), pages, specsDir, outDir, opts...)
}

func GeneratePage(page parser.PageSpec, specsDir string, opts ...GenerateOptions) string {
    opt := DefaultOptions()
    if len(opts) > 0 { opt = mergeOpt(opt, opts[0]) }
    return DefaultTarget().GeneratePage(page, specsDir, opt)
}

func GenerateWith(t Target, pages []parser.PageSpec, specsDir, outDir string, opts ...GenerateOptions) (*GenerateResult, error) {
    opt := DefaultOptions()
    if len(opts) > 0 { opt = mergeOpt(opt, opts[0]) }
    os.MkdirAll(outDir, 0o755)
    for _, page := range pages {
        code := t.GeneratePage(page, specsDir, opt)
        path := filepath.Join(outDir, page.Name + t.FileExtension())
        os.WriteFile(path, []byte(code), 0o644)
    }
    return &GenerateResult{Pages: len(pages), Dependencies: t.Dependencies(pages)}, nil
}
```

---

## Step 4: ReactTarget.Dependencies 구현

기존 Generate() 내부의 의존성 수집 로직을 ReactTarget.Dependencies로 이동:

```go
func (r *ReactTarget) Dependencies(pages []parser.PageSpec) map[string]string {
    deps := map[string]string{}
    for _, page := range pages {
        is := collectImports(page, "")
        if is.useQuery || is.useMutation || is.useQueryClient {
            deps["@tanstack/react-query"] = "^5"
        }
        if is.useForm { deps["react-hook-form"] = "^7" }
        if is.useParams { deps["react-router-dom"] = "^6" }
    }
    return deps
}
```

---

## Step 5: 인터페이스 컴파일 검증 + 테스트

1. `var _ Target = (*ReactTarget)(nil)` 컴파일 검증 추가 (target.go)
2. 기존 generator_test.go — 변경 없이 전부 통과 확인
3. 추가 테스트: `TestGenerateWith` — `GenerateWith(DefaultTarget(), ...)` 결과가 `Generate(...)` 결과와 동일한지 확인

```bash
go test ./... -count=1
```

---

## 파일별 변경 요약

| 파일 | 유형 | 예상 줄 수 |
|---|---|---|
| `generator/target.go` | NEW | ~25 |
| `generator/react_target.go` | NEW (generator.go에서 이동) | ~120 |
| `generator/react_imports.go` | RENAME (imports.go) | 145 (변경 없음) |
| `generator/react_templates.go` | RENAME (templates.go) | 521 (변경 없음) |
| `generator/generator.go` | 수정 (위임 래퍼로 축소) | ~120 |
| `generator/generator_test.go` | 테스트 1개 추가 | +15 |

## 위험도

**낮음** — 내부 리팩토링. 외부 API 시그니처 불변. 기존 테스트가 안전망.

## 수행 순서

1. `target.go` 생성 (인터페이스 정의)
2. `imports.go` → `react_imports.go` 이름 변경
3. `templates.go` → `react_templates.go` 이름 변경
4. `react_target.go` 생성 (generator.go에서 ReactTarget 로직 추출)
5. `generator.go` 축소 (위임 래퍼 + 공통 유틸만 유지)
6. 컴파일 확인
7. 테스트 전체 통과 확인
8. 추가 테스트 작성
