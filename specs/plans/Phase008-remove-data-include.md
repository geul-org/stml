✅ 완료

# Phase 8: data-include 제거

## 목표

`data-include` 속성을 STML에서 완전히 제거한다.

## 근거

`x-include`는 model/SQL 계층(fullend)을 위한 선언이며, STML 프론트엔드 선언과는 무관하다. `data-include`는 UI에 아무 흔적도 남기지 않고 API 호출 파라미터만 바꾸는 인프라 관심사이므로 SSOT 원칙에 맞지 않는다. 이 제거로 수정지시서004의 `x-include` 런타임 이름 변환 문제도 해소된다.

## 변경 파일 목록

| 파일 | 변경 유형 | 내용 |
|---|---|---|
| `parser/types.go` | 수정 | `FetchBlock.Includes` 필드 제거 |
| `parser/parser.go` | 수정 | `data-include` 파싱 코드 제거 |
| `parser/parser_test.go` | 수정 | `data-include` 관련 테스트 제거 |
| `generator/react_templates.go` | 수정 | `renderInfraApiArgs`에서 include 처리 제거, `hasInfra` 조건에서 Includes 제거 |
| `generator/generator_test.go` | 수정 | `data-include="author"` 및 `include: 'author'` assertion 제거 |
| `validator/symbol.go` | 수정 | `IncludeExt` 타입, `yamlIncludeExt` 타입, x-include 파싱 코드 제거 |
| `validator/validator.go` | 수정 | `data-include` 검증 코드 제거 |
| `validator/errors.go` | 수정 | `errIncludeNotAllowed` 함수 제거 |
| `validator/validator_test.go` | 수정 | `TestValidateIncludeNotAllowed` 테스트 제거, 기존 테스트에서 Includes 참조 제거 |

## 의존성

없음.

## 검증 방법

1. `go build ./...` 컴파일 통과
2. `go test ./... -count=1` 전체 통과
3. `data-include` / `Includes` / `IncludeExt` 문자열이 Go 소스에 남아있지 않음
