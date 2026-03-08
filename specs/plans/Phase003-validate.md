# Phase 3: Validate

> 파싱된 data-* 속성이 OpenAPI spec 및 custom.ts와 정합한지 교차 검증한다.

## 목표

Phase 1의 산출물(`[]PageSpec`)이 참조하는 operationId, 필드, 파라미터가 실제 OpenAPI spec에 존재하는지 심볼릭 검증한다. 코드 생성(Phase 2) 전에 실행하여 불일치를 조기 발견한다.

## 입력

1. Phase 1의 `[]PageSpec`
2. 심볼 테이블 소스:
   - OpenAPI spec — operationId, request/response schema, parameters
   - custom.ts — 프론트 계산 함수명

## 산출

- 검증 성공: exit 0, 코드젠 진행 가능
- 검증 실패: 에러 목록 출력, exit 1

## 검증 항목

### 1. operationId 존재 검증

`data-fetch="ListSessions"` → OpenAPI에 operationId `ListSessions`이 존재하는가?

```
ERROR: session-page.html — data-fetch="ListSessions": OpenAPI에 "ListSessions" operationId가 없습니다
```

`data-action="CreateSession"` → OpenAPI에 operationId `CreateSession`이 존재하는가?

```
ERROR: session-page.html — data-action="CreateSession": OpenAPI에 "CreateSession" operationId가 없습니다
```

### 2. HTTP 메서드 검증

`data-fetch`의 operationId는 GET 메서드여야 한다.
`data-action`의 operationId는 POST/PUT/DELETE 메서드여야 한다.

```
ERROR: session-page.html — data-fetch="CreateSession": "CreateSession"은 POST 메서드입니다 (GET이어야 함)
```

### 3. 파라미터 검증

`data-param-project-id` → 해당 operationId의 parameters에 `projectId`가 존재하는가?

```
ERROR: session-page.html — data-param-project-id: "ListSessions"의 parameters에 "projectId"가 없습니다
```

### 4. request 필드 검증

`data-field="Command"` → 해당 operationId의 request body schema에 `Command` 필드가 존재하는가?

```
ERROR: session-page.html — data-field="Command": "CreateSession"의 request schema에 "Command" 필드가 없습니다
```

### 5. response 필드 검증

`data-bind="status"` → 해당 operationId의 response schema에 `status` 필드가 존재하는가?

OpenAPI response에 없으면 custom.ts에서 찾는다. 둘 다 없으면 에러.

```
ERROR: project-detail-page.html — data-bind="status": "GetProject"의 response schema에도, custom.ts에도 "status"가 없습니다
```

### 6. 배열 타입 검증

`data-each="sessions"` → 해당 operationId의 response schema에서 `sessions`가 배열 타입인가?

```
ERROR: session-page.html — data-each="sessions": "ListSessions"의 response에서 "sessions"는 배열이 아닙니다
```

### 7. 컴포넌트 존재 검증

`data-component="Chart"` → `frontend/components/Chart.tsx`가 존재하는가?

```
ERROR: project-detail-page.html — data-component="Chart": frontend/components/Chart.tsx 파일이 없습니다
```

### 8. custom.ts 함수 검증

data-bind 필드가 OpenAPI response에 없고 custom.ts 함수로 해석될 때, 해당 함수가 실제로 export되어 있는가?

```
ERROR: cart-page.html — data-bind="totalPrice": OpenAPI response에 없고, cart-page.custom.ts에 "totalPrice" 함수도 없습니다
```

## 심볼 테이블 구성

### OpenAPI spec에서 수집

```go
type APISymbol struct {
    Method         string            // "GET", "POST", "PUT", "DELETE"
    Parameters     []ParamSymbol     // path/query 파라미터
    RequestFields  map[string]string // "Command" → "string"
    ResponseFields map[string]FieldSymbol // "sessions" → {Type: "array", ItemType: "Session"}
}

type ParamSymbol struct {
    Name string // "projectId"
    In   string // "path", "query"
}

type FieldSymbol struct {
    Type     string // "string", "integer", "array", "object"
    ItemType string // 배열일 때 항목 타입
}

type SymbolTable struct {
    Operations map[string]APISymbol // operationId → APISymbol
}
```

소스: OpenAPI spec(yaml) 파싱 (`gopkg.in/yaml.v3`)

### custom.ts에서 수집

```go
type CustomSymbol struct {
    Functions []string // export된 함수명 리스트 (e.g. ["totalPrice", "discountedPrice"])
}
```

소스: custom.ts 파일을 텍스트로 읽어 `export function <name>` 패턴 추출

### 컴포넌트 파일 존재 확인

소스: `<root>/frontend/components/<Name>.tsx` 파일 존재 여부 (os.Stat)

## 구현 계획

### 1. 디렉토리 구조

```
artifacts/
  internal/
    validator/
      validator.go       # 검증 메인 로직
      symbol.go          # 심볼 테이블 구성
      errors.go          # 에러 타입 정의
    validator_test.go    # 테스트
```

### 2. 심볼 테이블 로더

- OpenAPI 로더: yaml 파싱 → paths 순회 → operationId별 APISymbol 구성
- custom.ts 로더: 텍스트 스캔 → `export function` 패턴 매칭
- 컴포넌트 로더: 파일 시스템 확인

### 3. 검증 엔진

```
심볼 테이블 구성
  → PageSpec 순회
    → FetchBlock 순회
      → operationId 존재 + GET 메서드 검증
      → param 검증
      → bind 필드 검증 (response → custom.ts fallback)
      → each 배열 타입 검증
      → component 존재 검증
    → ActionBlock 순회
      → operationId 존재 + POST/PUT/DELETE 메서드 검증
      → param 검증
      → field 검증
  → 에러 리포트 출력
```

### 4. CLI 연동

- `stml validate` 명령: parse → validate 파이프라인
- `stml gen`에서도 validate를 먼저 실행 (검증 실패 시 코드 생성 중단)

### 5. 테스트

- 정상 케이스: 기획서 예시가 검증 통과
- 실패 케이스: 존재하지 않는 operationId, 누락된 필드, 배열이 아닌 each, 없는 컴포넌트, 없는 custom 함수

## 완료 기준

- [ ] OpenAPI spec에서 심볼 테이블 구성
- [ ] custom.ts에서 함수명 추출
- [ ] 8가지 검증 항목 모두 구현
- [ ] `stml validate` 명령으로 검증 실행
- [ ] 에러 메시지에 파일명, 속성, 구체적 원인 포함
- [ ] `stml gen`이 validate 실패 시 코드 생성 중단
