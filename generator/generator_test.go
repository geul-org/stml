package generator

import (
	"strings"
	"testing"

	"github.com/geul-org/stml/parser"
)

func TestGenerateLoginPage(t *testing.T) {
	page, _ := parser.ParseReader("login-page.html", strings.NewReader(`<main class="flex items-center justify-center min-h-screen">
  <div data-action="Login" class="space-y-4">
    <input data-field="Email" type="email" placeholder="이메일" class="w-full px-3 py-2 border rounded" />
    <input data-field="Password" type="password" placeholder="비밀번호" class="w-full px-3 py-2 border rounded" />
    <button type="submit">로그인</button>
  </div>
</main>`))

	code := GeneratePage(page, "")

	assertContains(t, code, "export default function LoginPage()")
	assertContains(t, code, "useMutation")
	assertContains(t, code, "api.Login")
	assertContains(t, code, `placeholder="이메일"`)
	assertContains(t, code, `placeholder="비밀번호"`)
	assertContains(t, code, `className="w-full px-3 py-2 border rounded"`)
	assertContains(t, code, `type="email"`)
	assertContains(t, code, `type="password"`)
	assertContains(t, code, ">로그인</")
	assertNotContains(t, code, ">제출</")
	assertNotContains(t, code, "useQuery(")
	assertNotContains(t, code, "useParams")
}

func TestGenerateMyReservationsPage(t *testing.T) {
	page, _ := parser.ParseReader("my-reservations-page.html", strings.NewReader(`<main class="max-w-4xl mx-auto p-6">
  <section data-fetch="ListMyReservations" class="mb-8">
    <ul data-each="reservations" class="space-y-3">
      <li class="flex justify-between p-4 border rounded">
        <span data-bind="RoomID" class="font-semibold"></span>
        <span data-bind="Status" class="px-2 py-1 text-sm rounded bg-gray-100"></span>
      </li>
    </ul>
    <p data-state="reservations.empty" class="text-gray-400">예약이 없습니다</p>
  </section>
  <div data-action="CreateReservation" class="space-y-4">
    <input data-field="RoomID" type="number" placeholder="스터디룸 번호" class="w-full px-3 py-2 border rounded" />
    <div data-component="DatePicker" data-field="StartAt" />
    <div data-component="DatePicker" data-field="EndAt" />
    <button type="submit">예약하기</button>
  </div>
</main>`))

	code := GeneratePage(page, "")

	assertContains(t, code, "export default function MyReservationsPage()")
	assertContains(t, code, "useQuery")
	assertContains(t, code, "api.ListMyReservations")
	// Tailwind classes
	assertContains(t, code, `className="mb-8"`)
	assertContains(t, code, `className="space-y-3"`)
	assertContains(t, code, `className="font-semibold"`)
	// Each with proper tags
	assertContains(t, code, "<ul")
	assertContains(t, code, "<li")
	// State with text
	assertContains(t, code, "예약이 없습니다")
	assertContains(t, code, "length === 0")
	// Action
	assertContains(t, code, "api.CreateReservation")
	assertContains(t, code, `placeholder="스터디룸 번호"`)
	assertContains(t, code, ">예약하기</")
	assertContains(t, code, "import DatePicker from '@/components/DatePicker'")
	assertContains(t, code, "<DatePicker")
	// onSuccess
	assertContains(t, code, "queryKey: ['ListMyReservations']")
}

func TestGenerateReservationDetailPage(t *testing.T) {
	page, _ := parser.ParseReader("reservation-detail-page.html", strings.NewReader(`<main class="max-w-2xl mx-auto p-6">
  <article data-fetch="GetReservation" data-param-reservation-id="route.ReservationID">
    <span data-bind="reservation.Status" class="px-3 py-1 text-sm rounded bg-gray-100"></span>
    <dd data-bind="reservation.RoomID" class="font-semibold"></dd>
    <footer data-state="canCancel" class="mt-8 pt-4 border-t">
      <button data-action="CancelReservation" data-param-reservation-id="route.ReservationID">
        예약 취소
      </button>
    </footer>
  </article>
</main>`))

	code := GeneratePage(page, "")

	assertContains(t, code, "export default function ReservationDetailPage()")
	assertContains(t, code, "useParams")
	assertContains(t, code, "ReservationID")
	assertContains(t, code, "api.GetReservation")
	assertContains(t, code, `className="px-3 py-1 text-sm rounded bg-gray-100"`)
	assertContains(t, code, `className="font-semibold"`)
	assertContains(t, code, "<article")
	// data-state canCancel renders condition
	assertContains(t, code, ".canCancel")
	assertContains(t, code, `className="mt-8 pt-4 border-t"`)
	assertContains(t, code, "<footer")
	// CancelReservation as button (no fields)
	assertContains(t, code, "onClick")
	assertContains(t, code, "cancelReservationMutation")
}

func TestGenerateRoomEditPage(t *testing.T) {
	page, _ := parser.ParseReader("room-edit-page.html", strings.NewReader(`<main class="max-w-2xl mx-auto p-6">
  <div data-action="UpdateRoom" data-param-room-id="route.RoomID" class="space-y-4">
    <input data-field="Name" placeholder="스터디룸 이름" class="w-full px-3 py-2 border rounded" />
    <input data-field="Capacity" type="number" placeholder="수용 인원" class="w-full px-3 py-2 border rounded" />
    <input data-field="Location" placeholder="위치" class="w-full px-3 py-2 border rounded" />
    <button type="submit">수정</button>
  </div>
  <footer data-state="canDelete" class="mt-8 pt-4 border-t">
    <button data-action="DeleteRoom" data-param-room-id="route.RoomID">
      스터디룸 삭제
    </button>
  </footer>
</main>`))

	code := GeneratePage(page, "")

	assertContains(t, code, "export default function RoomEditPage()")
	assertContains(t, code, "useParams")
	assertContains(t, code, "RoomID")
	assertContains(t, code, "api.UpdateRoom")
	assertContains(t, code, "api.DeleteRoom")
	assertContains(t, code, `placeholder="스터디룸 이름"`)
	assertContains(t, code, `placeholder="수용 인원"`)
	assertContains(t, code, `placeholder="위치"`)
	assertContains(t, code, ">수정</")
	// DeleteRoom: button onClick (no form)
	assertContains(t, code, "onClick")
	assertContains(t, code, "deleteRoomMutation.mutate")
	assertNotContains(t, code, "deleteRoomForm")
}

func TestToComponentName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"login-page", "LoginPage"},
		{"my-reservations-page", "MyReservationsPage"},
		{"room-edit-page", "RoomEditPage"},
	}
	for _, tt := range tests {
		got := toComponentName(tt.in)
		if got != tt.want {
			t.Errorf("toComponentName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGenerateWithInfraParams(t *testing.T) {
	page, _ := parser.ParseReader("list-page.html", strings.NewReader(`<main>
  <section data-fetch="ListItems"
           data-paginate
           data-sort="name:desc"
           data-filter="status,category"
>
    <ul data-each="items">
      <li><span data-bind="name"></span></li>
    </ul>
  </section>
</main>`))

	code := GeneratePage(page, "")

	// useState import
	assertContains(t, code, "import React, { useState } from 'react'")

	// useState hooks
	assertContains(t, code, "const [page, setPage] = useState(1)")
	assertContains(t, code, "const [limit] = useState(20)")
	assertContains(t, code, "const [sortBy, setSortBy] = useState('name')")
	assertContains(t, code, "const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')")
	assertContains(t, code, "const [filters, setFilters] = useState<Record<string, string>>({})")

	// queryKey includes infra params
	assertContains(t, code, "page, limit")
	assertContains(t, code, "sortBy, sortDir")
	assertContains(t, code, "filters")

	// Filter UI
	assertContains(t, code, `placeholder="status"`)
	assertContains(t, code, `placeholder="category"`)
	assertContains(t, code, "setFilters")

	// Sort UI
	assertContains(t, code, "setSortBy")
	assertContains(t, code, "setSortDir")

	// Pagination UI
	assertContains(t, code, "setPage")
	assertContains(t, code, ">이전</button>")
	assertContains(t, code, ">다음</button>")
}

func TestGenerateWithoutInfraParams(t *testing.T) {
	// Ensure existing pages without infra params still work
	page, _ := parser.ParseReader("simple-page.html", strings.NewReader(`<section data-fetch="GetItem">
  <span data-bind="name"></span>
</section>`))

	code := GeneratePage(page, "")

	assertNotContains(t, code, "useState")
	assertNotContains(t, code, "setPage")
	assertNotContains(t, code, "setSortBy")
	assertNotContains(t, code, "setFilters")
	assertNotContains(t, code, "이전")
	assertNotContains(t, code, "다음")
}

func TestGenerateOptionsAPIImportPath(t *testing.T) {
	page, _ := parser.ParseReader("login-page.html", strings.NewReader(`<main>
  <div data-action="Login">
    <input data-field="Email" type="email" />
    <button type="submit">로그인</button>
  </div>
</main>`))

	opts := GenerateOptions{APIImportPath: "../api", UseClient: false}
	code := GeneratePage(page, "", opts)

	assertContains(t, code, `import { api } from '../api'`)
	assertNotContains(t, code, `@/lib/api`)
	assertNotContains(t, code, "'use client'")
}

func TestGenerateOptionsDefaults(t *testing.T) {
	page, _ := parser.ParseReader("login-page.html", strings.NewReader(`<main>
  <div data-action="Login">
    <input data-field="Email" type="email" />
    <button type="submit">로그인</button>
  </div>
</main>`))

	// No opts — should use defaults
	code := GeneratePage(page, "")

	assertContains(t, code, `import { api } from '@/lib/api'`)
	assertContains(t, code, "'use client'")
}

func TestGenerateResultDependencies(t *testing.T) {
	page, _ := parser.ParseReader("login-page.html", strings.NewReader(`<main>
  <div data-action="Login">
    <input data-field="Email" type="email" />
    <button type="submit">로그인</button>
  </div>
</main>`))

	outDir := t.TempDir()
	result, err := Generate([]parser.PageSpec{page}, "", outDir)
	if err != nil {
		t.Fatal(err)
	}

	if result.Pages != 1 {
		t.Errorf("expected 1 page, got %d", result.Pages)
	}
	if result.Dependencies["@tanstack/react-query"] != "^5" {
		t.Errorf("expected @tanstack/react-query ^5, got %q", result.Dependencies["@tanstack/react-query"])
	}
	if result.Dependencies["react-hook-form"] != "^7" {
		t.Errorf("expected react-hook-form ^7, got %q", result.Dependencies["react-hook-form"])
	}
}

func TestGenerateWithDefaultTarget(t *testing.T) {
	page, _ := parser.ParseReader("login-page.html", strings.NewReader(`<main>
  <div data-action="Login">
    <input data-field="Email" type="email" />
    <button type="submit">로그인</button>
  </div>
</main>`))

	outDir1 := t.TempDir()
	outDir2 := t.TempDir()

	r1, err := Generate([]parser.PageSpec{page}, "", outDir1)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := GenerateWith(DefaultTarget(), []parser.PageSpec{page}, "", outDir2)
	if err != nil {
		t.Fatal(err)
	}

	if r1.Pages != r2.Pages {
		t.Errorf("Pages mismatch: %d vs %d", r1.Pages, r2.Pages)
	}
	for k, v := range r1.Dependencies {
		if r2.Dependencies[k] != v {
			t.Errorf("Dependency %s: %q vs %q", k, v, r2.Dependencies[k])
		}
	}
}

func TestStateLoadingErrorVarName(t *testing.T) {
	page, _ := parser.ParseReader("course-list-page.html", strings.NewReader(`<main>
  <section data-fetch="ListCourses">
    <ul data-each="courses">
      <li><span data-bind="name"></span></li>
    </ul>
    <div data-state="courses.loading" class="text-gray-500">로딩 중...</div>
    <div data-state="courses.error" class="text-red-500">불러오지 못했습니다</div>
  </section>
</main>`))

	code := GeneratePage(page, "")

	// useQuery defines: listCoursesDataLoading, listCoursesDataError
	assertContains(t, code, "listCoursesDataLoading")
	assertContains(t, code, "listCoursesDataError")
	// Must NOT generate old-style short variable names
	assertNotContains(t, code, "coursesLoading")
	assertNotContains(t, code, "coursesError")
}

func assertContains(t *testing.T, code, substr string) {
	t.Helper()
	if !strings.Contains(code, substr) {
		t.Errorf("generated code does not contain %q\n--- code ---\n%s", substr, code)
	}
}

func assertNotContains(t *testing.T, code, substr string) {
	t.Helper()
	if strings.Contains(code, substr) {
		t.Errorf("generated code should not contain %q", substr)
	}
}
