# STML — fullend 연동을 위한 작업지시서

## 배경

fullend CLI는 stml을 Go 모듈로 import하여 parser, validator, generator를 라이브러리로 호출해야 한다.
현재 stml의 모든 패키지가 `artifacts/internal/` 아래에 있어 외부 모듈에서 import 불가능하다.

## 작업 1: 패키지 공개 경로 재배치

### 변경 내용

```
변경 전:
artifacts/internal/parser/
artifacts/internal/validator/
artifacts/internal/generator/

변경 후:
parser/
validator/
generator/
```

모듈 루트 바로 아래로 이동한다. `artifacts/internal/` 경로를 제거한다.

### import 경로 변경

```
변경 전:
github.com/park-jun-woo/stml/artifacts/internal/parser
github.com/park-jun-woo/stml/artifacts/internal/validator
github.com/park-jun-woo/stml/artifacts/internal/generator

변경 후:
github.com/park-jun-woo/stml/parser
github.com/park-jun-woo/stml/validator
github.com/park-jun-woo/stml/generator
```

### 수정 대상 파일

1. `artifacts/cmd/stml/main.go` — import 경로 3개 변경
2. `validator/validator.go` — parser 패키지 import 경로 변경
3. `generator/generator.go` — parser 패키지 import 경로 변경
4. `generator/templates.go` — parser 패키지 import 있으면 변경

### 확인

- `go build ./...` 성공
- `go test ./...` 전체 통과
- 기존 테스트 파일의 import 경로도 모두 변경

## 작업 2: cmd 경로 정리

CLI 엔트리포인트도 함께 정리한다.

```
변경 전: artifacts/cmd/stml/main.go
변경 후: cmd/stml/main.go
```

## 작업 후 디렉토리 구조

```
stml/
├── cmd/stml/main.go            # CLI 엔트리포인트
├── parser/
│   ├── types.go                # PageSpec, FetchBlock, ActionBlock, ...
│   ├── parser.go               # ParseDir, ParseFile, ParseReader
│   └── parser_test.go
├── validator/
│   ├── errors.go               # ValidationError
│   ├── validator.go            # Validate
│   ├── symbol.go               # SymbolTable, LoadOpenAPI, LoadCustomTS
│   └── validator_test.go
├── generator/
│   ├── generator.go            # Generate, GeneratePage
│   ├── templates.go            # 내부 렌더링 헬퍼
│   └── generator_test.go
├── specs/                      # 예제 SSOT
├── files/                      # 문서
├── go.mod
└── go.sum
```

## 공개 API (변경 없음)

패키지 이동만 하고 API는 그대로 유지한다. 코드 변경 없음.

### parser
- `ParseDir(dir string) ([]PageSpec, error)`
- `ParseFile(path string) (PageSpec, error)`
- `ParseReader(filename string, r io.Reader) (PageSpec, error)`
- 타입: `PageSpec`, `FetchBlock`, `ActionBlock`, `FieldBind`, `ParamBind`, `EachBlock`, `StateBind`, `ComponentRef`, `StaticElement`, `ChildNode`, `SortDecl`

### validator
- `Validate(pages []parser.PageSpec, projectRoot string) []ValidationError`
- `LoadOpenAPI(path string) (*SymbolTable, error)`
- `LoadCustomTS(path string) (*CustomSymbol, error)`
- 타입: `SymbolTable`, `APISymbol`, `ParamSymbol`, `FieldSymbol`, `PaginationExt`, `SortExt`, `FilterExt`, `IncludeExt`, `CustomSymbol`, `ValidationError`

### generator
- `Generate(pages []parser.PageSpec, specsDir, outDir string) error`
- `GeneratePage(page parser.PageSpec, specsDir string) string`

## 주의사항

- `artifacts/` 디렉토리에는 test fixtures와 manual 문서만 남긴다
- `go.mod`의 module path `github.com/park-jun-woo/stml`은 변경하지 않는다
- 빈 `artifacts/internal/` 디렉토리는 삭제한다
