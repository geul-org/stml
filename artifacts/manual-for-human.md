# STML 사용자 매뉴얼

## 개요

STML(SSOT Template Markup Language)은 HTML5 `data-*` 속성으로 프론트엔드 UI와 API 바인딩을 선언하고, React TSX 컴포넌트를 자동 생성하는 Go CLI 도구입니다.

**핵심 아이디어:** HTML에 "무엇을 보여주고 무엇을 하는가"만 선언하면, "어떻게 렌더링하는가"는 코드젠이 채웁니다.

## 설치 및 실행

```bash
go install github.com/geul-org/stml/cmd/stml@latest
```

### CLI 명령어

```bash
# HTML 파싱 결과 확인 (JSON 출력)
stml parse specs/dummy-study/frontend

# OpenAPI 교차 검증
stml validate specs/dummy-study

# 검증 + TSX 코드 생성
stml gen specs/dummy-study artifacts/dummy-study/frontend
```

## 빠른 시작

### 1. OpenAPI 스펙 준비

```yaml
# specs/my-project/api/openapi.yaml
paths:
  /items:
    get:
      operationId: ListItems
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      $ref: '#/components/schemas/Item'
```

### 2. STML HTML 작성

```html
<!-- specs/my-project/frontend/item-list-page.html -->
<main class="max-w-4xl mx-auto p-6">

  <h1 class="text-2xl font-bold mb-6">아이템 목록</h1>

  <section data-fetch="ListItems">
    <ul data-each="items" class="space-y-2">
      <li class="p-3 border rounded">
        <span data-bind="name" class="font-semibold"></span>
        <span data-bind="status" class="text-sm text-gray-500"></span>
      </li>
    </ul>
    <p data-state="items.empty" class="text-gray-400 py-8">아이템이 없습니다</p>
  </section>

</main>
```

### 3. 코드 생성

```bash
stml gen specs/my-project artifacts/my-project/frontend
```

생성 결과: `artifacts/my-project/frontend/item-list-page.tsx`

## data-* 속성 가이드

### 데이터 조회: `data-fetch`

GET 엔드포인트를 연결합니다. `useQuery` 훅으로 변환됩니다.

```html
<article data-fetch="GetProject" data-param-project-id="route.projectId">
  <h1 data-bind="name"></h1>
  <p data-bind="description"></p>
</article>
```

### 데이터 변경: `data-action`

POST/PUT/DELETE 엔드포인트를 연결합니다.

**폼이 있는 경우** (data-field 포함):
```html
<div data-action="CreateItem" class="space-y-4">
  <input data-field="Name" placeholder="이름" class="border rounded px-3 py-2" />
  <input data-field="Price" type="number" placeholder="가격" />
  <button type="submit">생성</button>
</div>
```

**버튼만 있는 경우** (data-field 없음):
```html
<button data-action="DeleteItem" data-param-item-id="route.itemId"
        class="bg-red-500 text-white px-4 py-2 rounded">
  삭제
</button>
```

### 목록 반복: `data-each`

배열 필드를 반복합니다. 첫 번째 자식 요소가 반복 템플릿이 됩니다.

```html
<ul data-each="reservations" class="space-y-2">
  <li class="flex justify-between p-3 border rounded">
    <span data-bind="RoomID" class="font-semibold"></span>
    <span data-bind="Status"></span>
  </li>
</ul>
```

### 조건부 표시: `data-state`

```html
<!-- 목록이 비었을 때 -->
<p data-state="items.empty" class="text-gray-400">항목이 없습니다</p>

<!-- boolean 조건 -->
<footer data-state="canDelete" class="mt-8 border-t pt-4">
  <button data-action="DeleteRoom">삭제</button>
</footer>
```

### 커스텀 컴포넌트: `data-component`

```html
<!-- 입력 위임 -->
<div data-component="DatePicker" data-field="StartAt" />

<!-- 표시 위임 -->
<div data-component="Chart" data-bind="usageData" />
```

컴포넌트 래퍼 파일이 필요합니다:

```tsx
// specs/my-project/frontend/components/DatePicker.tsx
export { default } from 'react-datepicker'
```

### 파라미터: `data-param-*`

URL 경로 파라미터를 바인딩합니다. HTML5는 속성명을 소문자로 변환하므로 **kebab-case**를 사용합니다.

```html
<!-- data-param-reservation-id → ReservationID로 변환 -->
<article data-fetch="GetReservation" data-param-reservation-id="route.ReservationID">
```

## 인프라 파라미터 (페이지네이션, 정렬, 필터)

OpenAPI에 x- 확장이 선언된 목록 엔드포인트에 사용합니다.

### OpenAPI 준비

```yaml
/items:
  get:
    operationId: ListItems
    x-pagination:
      style: offset
      defaultLimit: 20
      maxLimit: 100
    x-sort:
      allowed: [name, created_at]
      default: name
      direction: asc
    x-filter:
      allowed: [status, category]
```

### HTML 선언

```html
<section data-fetch="ListItems"
         data-paginate
         data-sort="name:asc"
         data-filter="status,category">
  <ul data-each="items">
    <li><span data-bind="name"></span></li>
  </ul>
</section>
```

### 생성 결과

- `data-paginate` → `useState(page, limit)` + 이전/다음 버튼
- `data-sort` → `useState(sortBy, sortDir)` + 정렬 토글 버튼
- `data-filter` → `useState(filters)` + 필터 입력 UI
- `data-include` → API 호출 시 include 파라미터 전달

## 프론트엔드 계산: custom.ts

OpenAPI 응답에 없는 계산값은 동반 파일로 제공합니다:

```html
<!-- specs/frontend/cart-page.html -->
<span data-bind="totalPrice"></span>
```

```ts
// specs/frontend/cart-page.custom.ts
export function totalPrice(items) {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0)
}
```

## 검증

`stml validate`는 12가지 교차 검증을 실행합니다:

- operationId가 OpenAPI에 존재하는가
- HTTP 메서드가 올바른가 (fetch=GET, action=POST/PUT/DELETE)
- 파라미터가 OpenAPI parameters에 있는가
- 요청 필드가 request schema에 있는가
- 응답 필드가 response schema 또는 custom.ts에 있는가
- data-each 필드가 배열 타입인가
- 컴포넌트 .tsx 파일이 존재하는가
- 인프라 속성이 OpenAPI x- 확장의 allowed 목록에 포함되는가

## 프로젝트 구조

```
specs/
  <project>/
    api/openapi.yaml              ← OpenAPI 3.x (SSOT)
    frontend/
      <page-name>.html            ← STML 선언 (SSOT)
      <page-name>.custom.ts       ← 프론트 계산 (선택)
      components/<Name>.tsx       ← 컴포넌트 래퍼/구현

artifacts/
  <project>/
    frontend/<page-name>.tsx      ← 생성 결과 (수정 금지)
```

**SSOT 원칙:** `artifacts/`는 직접 수정하지 않습니다. `specs/`를 수정하고 `stml gen`으로 재생성합니다.

## 보존되는 것들

코드젠은 원본 HTML의 다음 요소를 보존합니다:

- HTML 태그 (section, article, ul, li, header, footer 등)
- Tailwind CSS 클래스
- input의 placeholder, type 속성
- 버튼 텍스트
- 정적 텍스트 콘텐츠
- DOM 순서
