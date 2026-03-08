# OpenAPI x- 확장 — 인프라 파라미터 선언

OpenAPI 엔드포인트에 페이지네이션, 정렬, 필터, 관계 포함 기능을 선언하는 4개의 x- 확장.

## x-pagination

목록 엔드포인트의 페이지네이션 방식을 선언한다.

| 필드 | 타입 | 필수 | 설명 |
|---|---|---|---|
| `style` | string | O | `offset` 또는 `cursor` |
| `defaultLimit` | integer | O | 기본 반환 건수 |
| `maxLimit` | integer | O | 최대 반환 건수 |

```yaml
x-pagination:
  style: offset
  defaultLimit: 20
  maxLimit: 100
```

- `style: offset` → response에 `total` 필드 권장
- `style: cursor` → response에 `nextCursor` 필드 권장

## x-sort

목록 엔드포인트의 정렬 가능 컬럼을 선언한다.

| 필드 | 타입 | 필수 | 설명 |
|---|---|---|---|
| `allowed` | string[] | O | 정렬 가능한 컬럼 목록 |
| `default` | string | X | 기본 정렬 컬럼 (없으면 allowed[0]) |
| `direction` | string | X | 기본 정렬 방향: `asc` 또는 `desc` (기본값: `asc`) |

```yaml
x-sort:
  allowed: [start_at, created_at, title]
  default: start_at
  direction: desc
```

## x-filter

목록 엔드포인트의 필터 가능 컬럼을 선언한다.

| 필드 | 타입 | 필수 | 설명 |
|---|---|---|---|
| `allowed` | string[] | O | 필터 가능한 컬럼 목록 |

```yaml
x-filter:
  allowed: [status, room_id]
```

검색은 필터의 특수 형태로 처리한다 (`?name=회의`). 모델에서 `LIKE` 또는 `tsvector`로 구현하는 것은 DDL/sqlc의 책임이다.

## x-include

목록/상세 엔드포인트에서 관계 리소스를 함께 반환할 수 있음을 선언한다.

| 필드 | 타입 | 필수 | 설명 |
|---|---|---|---|
| `allowed` | string[] | O | 포함 가능한 관계 리소스 목록 |

```yaml
x-include:
  allowed: [room, user]
```

## 복합 예시

```yaml
/api/reservations:
  get:
    operationId: ListReservations
    x-pagination:
      style: offset
      defaultLimit: 20
      maxLimit: 100
    x-sort:
      allowed: [start_at, created_at]
      default: start_at
      direction: desc
    x-filter:
      allowed: [status, room_id]
    x-include:
      allowed: [room, user]
    parameters:
      - name: userId
        in: query
        required: true
        schema:
          type: string
    responses:
      "200":
        content:
          application/json:
            schema:
              type: object
              properties:
                reservations:
                  type: array
                  items:
                    $ref: '#/components/schemas/Reservation'
                total:
                  type: integer
```

## 교차 검증

| 방향 | 검증 내용 |
|---|---|
| STML → OpenAPI x- | sort/filter/include 파라미터가 allowed에 포함되는가 |
| OpenAPI x- → DDL | allowed 컬럼이 테이블에 존재하는가 |
| SSaC → OpenAPI x- | x-가 있는 엔드포인트의 모델 호출에 QueryOpts가 포함되는가 |
| OpenAPI x- → response | pagination이면 배열 + total/cursor 필드가 있는가 |
