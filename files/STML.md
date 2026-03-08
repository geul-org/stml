# STML — SSOT Template Markup Language

> View 계층의 UI 구조와 데이터 바인딩을 선언적으로 정의하여, 프론트엔드의 SSOT를 구성한다.

## 빈 공간

기존 SSOT는 View 계층을 커버하지 않는다:

| 기존 SSOT | 커버 범위 | View 내부? |
|---|---|---|
| OpenAPI | API 경로, 파라미터, 응답 스키마 | X |
| SSaC | 서비스 함수 내부의 비즈니스 흐름 | X |
| SQL DDL | 테이블 구조, 인덱스, 제약 | X |

**"이 화면이 어떤 데이터를 보여주고, 어떤 액션을 트리거하는가"를 선언할 곳이 없다.** React 컴포넌트 구현 코드를 읽어야만 알 수 있다. 이 빈 공간을 STML이 채운다.

## 왜 가능한가 — UX 계약

목록은 스크롤한다, 폼은 제출한다, 버튼은 누른다, 삭제는 확인을 묻는다 — 수십 년간 인간에게 학습된 보편 UX 패턴이 존재한다. SSaC의 10개 타입이 서비스 계층의 보편 패턴이듯, STML의 `data-*` 8개가 이 UX 계약에 대응한다. 인간이 이 패턴을 기대하니까 선언만으로 코드젠이 가능하다.

## 왜 100%가 아닌가 — 커버율 60~70%

백엔드는 기계가 소비하므로 성공/실패 이진 계약으로 떨어진다. SSaC가 95%를 커버하는 이유다. 프론트엔드는 인간이 소비하므로 감각, 반응성, 맥락이 끼어든다. 드래그앤드롭, 실시간 협업, 제스처 인터랙션 같은 영역은 인류의 UX 합의가 아직 보편화되지 않았다. STML의 한계는 설계의 문제가 아니라, 선언 가능한 UX 패턴의 범위가 거기까지라는 뜻이다.

STML은 보편 UX 패턴 60~70%를 data-*로 선언하고, 나머지 30~40%는 커스텀 컴포넌트로 격리한다. 이것은 의도된 설계이며, 100% 커버를 목표로 하지 않는다.

## 핵심 개념

STML은 HTML5 + Tailwind CSS + `data-*` 속성으로 구성된다. 새 문법을 발명하지 않는다. AI가 이미 유창한 HTML을 그대로 사용하되, `data-*` 속성으로 OpenAPI 엔드포인트와의 바인딩을 선언한다.

**what(뭘 보여주고 뭘 하는가)만 선언하고, how(어떻게 렌더링하는가)는 코드젠이 채운다.**

## data-* 속성 체계

| 속성 | 역할 | 바인딩 대상 |
|---|---|---|
| `data-fetch` | GET 엔드포인트 연결 | OpenAPI operationId |
| `data-action` | POST/PUT/DELETE 엔드포인트 연결 | OpenAPI operationId |
| `data-field` | request 필드 바인딩 (입력) | OpenAPI request schema 필드 |
| `data-bind` | response 필드 바인딩 (출력) | OpenAPI response schema 필드 또는 custom.ts 함수 |
| `data-param-*` | path/query 파라미터 바인딩 | OpenAPI parameters |
| `data-each` | 배열 response 반복 | OpenAPI response 배열 필드 |
| `data-state` | 조건부 표시 | 로컬 UI 상태 |
| `data-component` | 커스텀 컴포넌트 위임 | React 컴포넌트명 |

## 3계층 컴포넌트 구조

STML은 표현 범위에 따라 세 계층으로 분리한다:

| 계층 | 커버율 | 표현 수단 | 예시 |
|---|---|---|---|
| STML data-* | 60~70% | 8개 고정 속성 | 목록, 폼, 조회, 생성, 삭제 |
| React 생태계 래퍼 | 20~30% | data-component + npm 래핑 | 달력, 에디터, 차트, 애니메이션 |
| 커스텀 컴포넌트 + JSDoc | 5~10% | data-component + 직접 구현 | 세상에 없는 UX 패턴 |

### 1계층 — STML data-* (60~70%)

보편 UX 패턴. data-* 속성으로 선언하고 심볼릭 코드젠으로 산출한다.

```html
<section data-fetch="ListSessions" data-param-project-id="currentProject.id">
  <ul data-each="sessions">
    <li>
      <span data-bind="command"></span>
      <span data-bind="status"></span>
    </li>
  </ul>
</section>

<div data-action="CreateSession">
  <input data-field="ProjectID" type="hidden" />
  <input data-field="Command" placeholder="명령어 입력" />
  <button type="submit">생성</button>
</div>
```

### 2계층 — React 생태계 래퍼 (20~30%)

검증된 라이브러리가 있는 특수 패턴. `data-component`로 위임하고 래퍼 파일로 연결한다.

```html
<div data-component="DatePicker" data-field="ReservedAt" />
<div data-component="RichEditor" data-field="Description" />
<div data-component="Chart" data-bind="usageData" />
<div data-component="PageTransition">
  <main data-fetch="GetProject">...</main>
</div>
```

래퍼 구현:

```tsx
// specs/frontend/components/DatePicker.tsx
export { default } from 'react-datepicker'

// specs/frontend/components/RichEditor.tsx
export { default } from '@tiptap/react'

// specs/frontend/components/Chart.tsx
export { BarChart as default } from 'recharts'

// specs/frontend/components/PageTransition.tsx
export { AnimatePresence as default } from 'framer-motion'

// specs/frontend/components/AnimatedList.tsx
export { default } from 'react-flip-move'
```

### 3계층 — 커스텀 컴포넌트 + JSDoc (5~10%)

React 생태계에도 없는 패턴. 직접 구현하되, JSDoc으로 요구사항을 SSOT에 보존한다. JSDoc이 이 컴포넌트의 SSOT이며, 요구사항이 바뀌면 JSDoc을 먼저 수정한다.

```html
<section data-fetch="ListCommits" data-param-pr-id="route.prId">
  <div data-component="CodeReviewTimeline" data-bind="commits" />
</section>
```

```tsx
// specs/frontend/components/CodeReviewTimeline.tsx

/**
 * 코드 리뷰 타임라인
 * - 커밋 단위로 세로 타임라인 표시
 * - 각 커밋에 인라인 코멘트 스레드 펼침/접힘
 * - 코멘트 작성 시 data-action="CreateReviewComment" 호출
 * - 리졸브 시 해당 스레드 흐리게 처리
 */
export default function CodeReviewTimeline({ data }) {
  // 구현
}
```

```html
<section data-fetch="ListTasks" data-param-board-id="route.boardId">
  <div data-component="KanbanBoard" data-bind="tasks" />
</section>
```

```tsx
// specs/frontend/components/KanbanBoard.tsx

/**
 * 칸반 보드
 * - 컬럼 간 드래그앤드롭으로 태스크 이동
 * - 이동 시 data-action="UpdateTaskStatus" 호출
 * - 낙관적 업데이트 적용
 * - 컬럼: todo, in-progress, done (서버 status enum 연동)
 */
export default function KanbanBoard({ data }) {
  // 구현
}
```

## 표현 예시 — 단순 (목록 + 생성)

```html
<!-- specs/frontend/session-page.html -->
<main class="flex flex-col h-screen">

  <header class="flex justify-between p-4 border-b">
    <h1 class="text-xl font-bold">세션 관리</h1>
  </header>

  <section data-fetch="ListSessions" data-param-project-id="currentProject.id" class="flex-1 p-4">
    <ul data-each="sessions" class="space-y-2">
      <li class="flex justify-between p-3 border rounded">
        <span data-bind="command" class="font-mono"></span>
        <span data-bind="status" class="text-sm text-gray-500"></span>
      </li>
    </ul>
  </section>

  <footer class="p-4 border-t">
    <div data-action="CreateSession" class="flex gap-2">
      <input data-field="ProjectID" type="hidden" />
      <input data-field="Command" placeholder="명령어 입력" class="flex-1 px-3 py-2 border rounded" />
      <button type="submit" class="px-4 py-2 bg-blue-500 text-white rounded">생성</button>
    </div>
  </footer>

</main>
```

## 표현 예시 — 복합 (조회 + 커스텀 + 삭제)

```html
<!-- specs/frontend/project-detail-page.html -->
<main class="max-w-4xl mx-auto p-6">

  <article data-fetch="GetProject" data-param-project-id="route.projectId">

    <header class="flex justify-between items-center mb-6">
      <h1 data-bind="name" class="text-2xl font-bold"></h1>
      <span data-bind="status" class="px-2 py-1 text-sm rounded bg-gray-100"></span>
    </header>

    <section class="mb-6">
      <p data-bind="description" class="text-gray-600"></p>
      <time data-bind="createdAt" class="text-sm text-gray-400"></time>
    </section>

    <section data-fetch="ListSessionsByProject" data-param-project-id="route.projectId">
      <h2 class="text-lg font-semibold mb-3">세션</h2>
      <ul data-each="sessions" class="space-y-2">
        <li class="flex justify-between p-3 border rounded">
          <span data-bind="command" class="font-mono"></span>
          <span data-bind="status"></span>
        </li>
      </ul>
      <p data-state="sessions.empty" class="text-gray-400">세션이 없습니다</p>
    </section>

    <section data-fetch="GetProjectStats" data-param-project-id="route.projectId" class="mt-6">
      <h2 class="text-lg font-semibold mb-3">통계</h2>
      <div data-component="Chart" data-bind="usageData" />
    </section>

    <footer data-state="canDelete" class="mt-8 pt-4 border-t">
      <button data-action="DeleteProject" data-param-project-id="route.projectId"
              class="px-4 py-2 bg-red-500 text-white rounded">
        프로젝트 삭제
      </button>
    </footer>

  </article>

</main>
```

## 프론트엔드 계산 — custom.ts 동반 파일

프론트엔드에서 즉각 반영해야 하는 계산 로직(장바구니 총액, 할인 적용 등)은 `.custom.ts` 동반 파일에 순수 함수로 작성한다. SSOT(`specs/`) 안에 위치하므로 사용자의 결정이 보존된다.

```html
<!-- specs/frontend/cart-page.html -->
<main class="max-w-2xl mx-auto p-6">

  <section data-fetch="GetCart">
    <ul data-each="items" class="space-y-2">
      <li class="flex justify-between p-3 border rounded">
        <span data-bind="name"></span>
        <span data-bind="price"></span>
        <input data-field="quantity" type="number" class="w-16 border rounded px-2" />
      </li>
    </ul>

    <div class="flex justify-between mt-4 pt-4 border-t font-bold">
      <span>총액</span>
      <span data-bind="totalPrice"></span>
    </div>
    <div class="flex justify-between text-green-600">
      <span>할인</span>
      <span data-bind="discountedPrice"></span>
    </div>
  </section>

</main>
```

```ts
// specs/frontend/cart-page.custom.ts

export function totalPrice(items) {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0)
}

export function discountedPrice(total, coupon) {
  return coupon ? total - coupon.amount : total
}
```

`data-bind`의 필드가 OpenAPI response에 없을 때, `.custom.ts`에 해당 이름의 함수가 있으면 프론트 계산값으로 처리된다. 둘 다 없으면 검증 에러.

## 정합성 검증

STML의 핵심 가치는 **OpenAPI와 프론트엔드 사이의 정합성을 배포 전에 심볼릭으로 잡는 것**이다.

```
                    OpenAPI (specs/api/)
                   ↙                    ↘
SSaC (specs/backend/service/)    STML (specs/frontend/)
  ← 백엔드 교차 검증               ← 프론트엔드 교차 검증
                   ↖
              DDL (specs/db/)
```

### OpenAPI ↔ STML

- `data-fetch="ListSessions"` → operationId가 OpenAPI에 존재하는가?
- `data-action="CreateSession"` → operationId가 존재하는가?
- `data-field="Command"` → 해당 엔드포인트 request schema에 `Command` 필드가 있는가?
- `data-bind="status"` → 해당 엔드포인트 response schema에 `status` 필드가 있는가?
- `data-param-project-id` → 해당 엔드포인트 parameters에 `projectId`가 있는가?
- `data-each="sessions"` → response schema에 `sessions`가 배열 타입인가?

### data-bind + custom.ts 검증

- `data-bind="totalPrice"` → OpenAPI response에 없음 → `.custom.ts`에 `totalPrice` 함수 있음 → 통과
- `data-bind="totalPrice"` → OpenAPI response에도 없고 `.custom.ts`에도 없음 → **에러**

### SSaC ↔ STML (간접 보장)

STML과 SSaC가 같은 OpenAPI operationId를 참조하므로, 양쪽 검증이 통과하면 프론트엔드가 호출하는 API와 백엔드가 처리하는 API가 일치함이 보장된다.

### 검증 파이프라인

```
1. stml validate          ← 심볼릭 (OpenAPI + custom.ts 교차 체크)
2. stml gen               ← 심볼릭 (data-* → React 훅 매핑) + custom.ts/component merge
```

검증도 코드젠도 심볼릭이다. data-* 속성 8개의 React 매핑이 전부 결정적이므로 SSaC와 마찬가지로 LLM 없이 템플릿 매칭으로 코드 생성이 가능하다.

## 설계 원칙

### 새 문법을 발명하지 않는다

STML은 유효한 HTML5 문서다. 브라우저에서 열면 Tailwind가 적용된 정적 마크업으로 보인다. `data-*` 속성은 HTML 표준이다. AI가 학습 데이터에서 이미 수십억 번 본 포맷이므로 추가 학습이 필요 없다.

### data-*로 표현 못하는 건 커스텀 컴포넌트로 위임한다

SSaC에서 sequence로 표현 못하는 로직을 `call`로 위임하듯, STML에서 `data-*`로 표현 못하는 UI를 `data-component`로 위임한다.

```
STML에서 표현 가능         → data-field, data-bind, data-each 등 (60~70%)
기본 트랜지션              → Tailwind 클래스 (transition-all, duration-300 등)
특수 입력 (달력, 에디터)    → data-component + React 생태계 래퍼 (20~30%)
애니메이션 (페이지 전환 등) → data-component + framer-motion 등 래퍼
프론트 계산 (총액, 할인)    → custom.ts (순수 함수)
세상에 없는 UX 패턴        → data-component + JSDoc 명세 + 직접 구현 (5~10%)
```

### SSOT 바깥에 사용자 결정이 존재하면 안 된다

코드젠이 산출하는 `artifacts/`는 언제든 재생성할 수 있다. 사용자의 결정은 `specs/` 안에만 존재한다:

```
specs/frontend/
  session-page.html          ← STML 선언 (what)
  cart-page.html             ← STML 선언
  cart-page.custom.ts        ← 프론트 계산 로직 (순수 함수)
  components/                ← 커스텀 컴포넌트
    DatePicker.tsx           ←   2계층: React 생태계 래퍼
    RichEditor.tsx           ←   2계층: React 생태계 래퍼
    Chart.tsx                ←   2계층: React 생태계 래퍼
    PageTransition.tsx       ←   2계층: React 생태계 래퍼
    AnimatedList.tsx         ←   2계층: React 생태계 래퍼
    KanbanBoard.tsx          ←   3계층: JSDoc 명세 + 직접 구현
    CodeReviewTimeline.tsx   ←   3계층: JSDoc 명세 + 직접 구현
```

### SSaC와 대칭 구조

| | SSaC | STML |
|---|---|---|
| 대상 | Service 계층 (백엔드) | View 계층 (프론트엔드) |
| 커버율 | ~95% | 60~70% |
| 선언 단위 | sequence (10개 고정 타입) | data-* (8개 고정 속성) |
| 위임 | call @func / @component | data-component / custom.ts |
| 참조 SSOT | OpenAPI + DDL | OpenAPI |
| 코드젠 | 심볼릭 (템플릿 매칭) | 심볼릭 (data-* → React 훅 매핑) |
| 검증 | 심볼릭 (교차 체크) | 심볼릭 (교차 체크) |
| 코드젠 결과 | Go 서비스 함수 | React 컴포넌트 |

### 페이지 단위 파일

파일 하나 = 페이지 하나. `data-component` 속성으로 컴포넌트 경계를 표시하되, 분리 여부는 코드젠이 결정한다.

## SSOT 계층 구조

```
OpenAPI          → Controller  "어떤 API가 있는가"        — openapi-generator
Model SSOT       → Model       "어떤 리소스가 있는가"
  ├─ SQL DDL     →   DB 모델                             — sqlc
  ├─ OpenAPI     →   외부 API                            — openapi-generator
  └─ Go interface →  비-DB 모델                           — 선언 자체가 코드
SSaC             → Service     "어떤 흐름으로 처리하는가"   — SSaC codegen (심볼릭)
STML             → View        "뭘 보여주고 뭘 하는가"     — STML codegen (심볼릭)
                   └─ 구현 코드  "어떻게 렌더링하는가"       (React, 상태관리, ...)
```

## 토큰 비용 비교

| | SSOT (STML) | 구현 코드 (React) |
|---|---|---|
| 줄 수 | 15~30줄 | 50~150줄 |
| 표현 내용 | 의도 (what) | 의도 + 구현 (what + how) |
| 상태 관리 | 없음 | useQuery, useState, mutation |
| 라이브러리 의존 | 없음 (커스텀만 래퍼) | React, TanStack Query, etc. |
| 변경 시점 | 화면 요구사항 변경 시 | 리팩토링, 라이브러리 교체 포함 |
