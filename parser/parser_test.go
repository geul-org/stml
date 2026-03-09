package parser

import (
	"strings"
	"testing"
)

func TestParseLoginPage(t *testing.T) {
	input := `<main>
  <div data-action="Login" class="space-y-4">
    <input data-field="Email" type="email" />
    <input data-field="Password" type="password" />
    <button type="submit">로그인</button>
  </div>
</main>`

	page, err := ParseReader("login-page.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if page.Name != "login-page" {
		t.Errorf("Name = %q, want %q", page.Name, "login-page")
	}
	if len(page.Fetches) != 0 {
		t.Errorf("Fetches = %d, want 0", len(page.Fetches))
	}
	if len(page.Actions) != 1 {
		t.Fatalf("Actions = %d, want 1", len(page.Actions))
	}

	action := page.Actions[0]
	if action.OperationID != "Login" {
		t.Errorf("OperationID = %q, want %q", action.OperationID, "Login")
	}
	if len(action.Fields) != 2 {
		t.Fatalf("Fields = %d, want 2", len(action.Fields))
	}
	assertField(t, action.Fields[0], "Email", "input", "email")
	assertField(t, action.Fields[1], "Password", "input", "password")
}

func TestParseMyReservationsPage(t *testing.T) {
	input := `<main>
  <section data-fetch="ListMyReservations">
    <ul data-each="reservations">
      <li>
        <span data-bind="RoomID"></span>
        <span data-bind="StartAt"></span>
        <span data-bind="EndAt"></span>
        <span data-bind="Status"></span>
      </li>
    </ul>
    <p data-state="reservations.empty">예약이 없습니다</p>
  </section>

  <div data-action="CreateReservation">
    <input data-field="RoomID" type="number" />
    <div data-component="DatePicker" data-field="StartAt" />
    <div data-component="DatePicker" data-field="EndAt" />
    <button type="submit">예약하기</button>
  </div>
</main>`

	page, err := ParseReader("my-reservations-page.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	// Fetch block
	if len(page.Fetches) != 1 {
		t.Fatalf("Fetches = %d, want 1", len(page.Fetches))
	}
	fetch := page.Fetches[0]
	if fetch.OperationID != "ListMyReservations" {
		t.Errorf("OperationID = %q, want %q", fetch.OperationID, "ListMyReservations")
	}

	// Each block
	if len(fetch.Eaches) != 1 {
		t.Fatalf("Eaches = %d, want 1", len(fetch.Eaches))
	}
	each := fetch.Eaches[0]
	if each.Field != "reservations" {
		t.Errorf("Each.Field = %q, want %q", each.Field, "reservations")
	}
	if len(each.Binds) != 4 {
		t.Errorf("Each.Binds = %d, want 4", len(each.Binds))
	}

	// State
	if len(fetch.States) != 1 {
		t.Fatalf("States = %d, want 1", len(fetch.States))
	}
	if fetch.States[0].Condition != "reservations.empty" {
		t.Errorf("State = %q, want %q", fetch.States[0].Condition, "reservations.empty")
	}

	// Action block
	if len(page.Actions) != 1 {
		t.Fatalf("Actions = %d, want 1", len(page.Actions))
	}
	action := page.Actions[0]
	if action.OperationID != "CreateReservation" {
		t.Errorf("OperationID = %q, want %q", action.OperationID, "CreateReservation")
	}
	if len(action.Fields) != 3 {
		t.Fatalf("Fields = %d, want 3", len(action.Fields))
	}
	assertField(t, action.Fields[0], "RoomID", "input", "number")
	assertField(t, action.Fields[1], "StartAt", "data-component:DatePicker", "")
	assertField(t, action.Fields[2], "EndAt", "data-component:DatePicker", "")
}

func TestParseReservationDetailPage(t *testing.T) {
	input := `<main>
  <article data-fetch="GetReservation" data-param-reservation-id="route.ReservationID">
    <span data-bind="reservation.Status"></span>
    <dd data-bind="reservation.RoomID"></dd>
    <dd data-bind="reservation.StartAt"></dd>
    <dd data-bind="reservation.EndAt"></dd>

    <footer data-state="canCancel">
      <button data-action="CancelReservation" data-param-reservation-id="route.ReservationID">
        예약 취소
      </button>
    </footer>
  </article>
</main>`

	page, err := ParseReader("reservation-detail-page.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Fetches) != 1 {
		t.Fatalf("Fetches = %d, want 1", len(page.Fetches))
	}
	fetch := page.Fetches[0]
	if fetch.OperationID != "GetReservation" {
		t.Errorf("OperationID = %q, want %q", fetch.OperationID, "GetReservation")
	}

	// Params
	if len(fetch.Params) != 1 {
		t.Fatalf("Params = %d, want 1", len(fetch.Params))
	}
	if fetch.Params[0].Name != "reservationId" {
		t.Errorf("Param.Name = %q, want %q", fetch.Params[0].Name, "reservationId")
	}
	if fetch.Params[0].Source != "route.ReservationID" {
		t.Errorf("Param.Source = %q, want %q", fetch.Params[0].Source, "route.ReservationID")
	}

	// Binds
	if len(fetch.Binds) != 4 {
		t.Errorf("Binds = %d, want 4", len(fetch.Binds))
	}

	// State
	if len(fetch.States) != 1 {
		t.Fatalf("States = %d, want 1", len(fetch.States))
	}
	if fetch.States[0].Condition != "canCancel" {
		t.Errorf("State = %q, want %q", fetch.States[0].Condition, "canCancel")
	}
}

func TestParseRoomEditPage(t *testing.T) {
	input := `<main>
  <div data-action="UpdateRoom" data-param-room-id="route.RoomID">
    <input data-field="Name" />
    <input data-field="Capacity" type="number" />
    <input data-field="Location" />
    <button type="submit">수정</button>
  </div>

  <footer data-state="canDelete">
    <button data-action="DeleteRoom" data-param-room-id="route.RoomID">
      스터디룸 삭제
    </button>
  </footer>
</main>`

	page, err := ParseReader("room-edit-page.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Actions) != 2 {
		t.Fatalf("Actions = %d, want 2", len(page.Actions))
	}

	update := page.Actions[0]
	if update.OperationID != "UpdateRoom" {
		t.Errorf("OperationID = %q, want %q", update.OperationID, "UpdateRoom")
	}
	if len(update.Params) != 1 {
		t.Fatalf("Params = %d, want 1", len(update.Params))
	}
	if update.Params[0].Name != "roomId" {
		t.Errorf("Param.Name = %q, want %q", update.Params[0].Name, "roomId")
	}
	if len(update.Fields) != 3 {
		t.Fatalf("Fields = %d, want 3", len(update.Fields))
	}

	del := page.Actions[1]
	if del.OperationID != "DeleteRoom" {
		t.Errorf("OperationID = %q, want %q", del.OperationID, "DeleteRoom")
	}
	if len(del.Params) != 1 {
		t.Fatalf("Params = %d, want 1", len(del.Params))
	}
}

func TestKebabToCamel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"project-id", "projectId"},
		{"ReservationID", "ReservationID"},
		{"room-id", "roomId"},
		{"a-b-c", "aBC"},
	}
	for _, tt := range tests {
		got := kebabToCamel(tt.in)
		if got != tt.want {
			t.Errorf("kebabToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPhase4_TagAndClassName(t *testing.T) {
	input := `<main class="max-w-4xl mx-auto p-6">
  <section data-fetch="ListMyReservations" class="mb-8">
    <ul data-each="reservations" class="space-y-3">
      <li class="flex justify-between p-4 border rounded">
        <span data-bind="RoomID" class="font-semibold"></span>
      </li>
    </ul>
  </section>
  <div data-action="CreateReservation" class="space-y-4">
    <input data-field="RoomID" type="number" placeholder="스터디룸 번호" class="w-full px-3 py-2 border rounded" />
    <button type="submit">예약하기</button>
  </div>
</main>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	// Fetch Tag + ClassName
	fetch := page.Fetches[0]
	if fetch.Tag != "section" {
		t.Errorf("Fetch.Tag = %q, want %q", fetch.Tag, "section")
	}
	if fetch.ClassName != "mb-8" {
		t.Errorf("Fetch.ClassName = %q, want %q", fetch.ClassName, "mb-8")
	}

	// Each Tag + ClassName + ItemTag + ItemClassName
	each := fetch.Eaches[0]
	if each.Tag != "ul" {
		t.Errorf("Each.Tag = %q, want %q", each.Tag, "ul")
	}
	if each.ClassName != "space-y-3" {
		t.Errorf("Each.ClassName = %q, want %q", each.ClassName, "space-y-3")
	}
	if each.ItemTag != "li" {
		t.Errorf("Each.ItemTag = %q, want %q", each.ItemTag, "li")
	}
	if each.ItemClassName != "flex justify-between p-4 border rounded" {
		t.Errorf("Each.ItemClassName = %q, want %q", each.ItemClassName, "flex justify-between p-4 border rounded")
	}

	// Bind ClassName
	if each.Binds[0].ClassName != "font-semibold" {
		t.Errorf("Bind.ClassName = %q, want %q", each.Binds[0].ClassName, "font-semibold")
	}

	// Action Tag + ClassName
	action := page.Actions[0]
	if action.Tag != "div" {
		t.Errorf("Action.Tag = %q, want %q", action.Tag, "div")
	}
	if action.ClassName != "space-y-4" {
		t.Errorf("Action.ClassName = %q, want %q", action.ClassName, "space-y-4")
	}

	// Field Placeholder + ClassName
	field := action.Fields[0]
	if field.Placeholder != "스터디룸 번호" {
		t.Errorf("Field.Placeholder = %q, want %q", field.Placeholder, "스터디룸 번호")
	}
	if field.ClassName != "w-full px-3 py-2 border rounded" {
		t.Errorf("Field.ClassName = %q, want %q", field.ClassName, "w-full px-3 py-2 border rounded")
	}

	// SubmitText
	if action.SubmitText != "예약하기" {
		t.Errorf("SubmitText = %q, want %q", action.SubmitText, "예약하기")
	}
}

func TestPhase4_StateText(t *testing.T) {
	input := `<main>
  <section data-fetch="ListMyReservations">
    <p data-state="reservations.empty" class="text-gray-400">예약이 없습니다</p>
  </section>
</main>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	state := page.Fetches[0].States[0]
	if state.Tag != "p" {
		t.Errorf("State.Tag = %q, want %q", state.Tag, "p")
	}
	if state.ClassName != "text-gray-400" {
		t.Errorf("State.ClassName = %q, want %q", state.ClassName, "text-gray-400")
	}
	if state.Text != "예약이 없습니다" {
		t.Errorf("State.Text = %q, want %q", state.Text, "예약이 없습니다")
	}
}

func TestPhase4_StateWithAction(t *testing.T) {
	input := `<main>
  <article data-fetch="GetReservation" data-param-reservation-id="route.ReservationID">
    <footer data-state="canCancel" class="mt-8 pt-4 border-t">
      <button data-action="CancelReservation" data-param-reservation-id="route.ReservationID">
        예약 취소
      </button>
    </footer>
  </article>
</main>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	state := page.Fetches[0].States[0]
	if state.Tag != "footer" {
		t.Errorf("State.Tag = %q, want %q", state.Tag, "footer")
	}
	if state.ClassName != "mt-8 pt-4 border-t" {
		t.Errorf("State.ClassName = %q, want %q", state.ClassName, "mt-8 pt-4 border-t")
	}
	if len(state.Children) != 1 {
		t.Fatalf("State.Children = %d, want 1", len(state.Children))
	}
	if state.Children[0].Kind != "action" {
		t.Errorf("State.Children[0].Kind = %q, want %q", state.Children[0].Kind, "action")
	}
	if state.Children[0].Action.OperationID != "CancelReservation" {
		t.Errorf("Action.OperationID = %q, want %q", state.Children[0].Action.OperationID, "CancelReservation")
	}
}

func TestPhase5_InfraParams(t *testing.T) {
	input := `<main>
  <section data-fetch="ListMyReservations"
           data-paginate
           data-sort="StartAt:desc"
           data-filter="Status,RoomID"
    <ul data-each="reservations">
      <li><span data-bind="RoomID"></span></li>
    </ul>
  </section>
</main>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	fetch := page.Fetches[0]

	// data-paginate
	if !fetch.Paginate {
		t.Error("Paginate = false, want true")
	}

	// data-sort
	if fetch.Sort == nil {
		t.Fatal("Sort = nil, want non-nil")
	}
	if fetch.Sort.Column != "StartAt" {
		t.Errorf("Sort.Column = %q, want %q", fetch.Sort.Column, "StartAt")
	}
	if fetch.Sort.Direction != "desc" {
		t.Errorf("Sort.Direction = %q, want %q", fetch.Sort.Direction, "desc")
	}

	// data-filter
	if len(fetch.Filters) != 2 {
		t.Fatalf("Filters = %d, want 2", len(fetch.Filters))
	}
	if fetch.Filters[0] != "Status" || fetch.Filters[1] != "RoomID" {
		t.Errorf("Filters = %v, want [Status RoomID]", fetch.Filters)
	}

}

func TestPhase5_SortDefaultDirection(t *testing.T) {
	input := `<section data-fetch="ListItems" data-sort="name">
  <ul data-each="items"><li><span data-bind="name"></span></li></ul>
</section>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	fetch := page.Fetches[0]
	if fetch.Sort == nil {
		t.Fatal("Sort = nil")
	}
	if fetch.Sort.Direction != "asc" {
		t.Errorf("Sort.Direction = %q, want %q", fetch.Sort.Direction, "asc")
	}
}

func TestPhase5_NoInfraParams(t *testing.T) {
	input := `<section data-fetch="GetItem">
  <span data-bind="name"></span>
</section>`

	page, err := ParseReader("test.html", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	fetch := page.Fetches[0]
	if fetch.Paginate {
		t.Error("Paginate = true, want false")
	}
	if fetch.Sort != nil {
		t.Error("Sort != nil, want nil")
	}
	if len(fetch.Filters) != 0 {
		t.Errorf("Filters = %d, want 0", len(fetch.Filters))
	}
}

func assertField(t *testing.T, f FieldBind, name, tag, typ string) {
	t.Helper()
	if f.Name != name {
		t.Errorf("Field.Name = %q, want %q", f.Name, name)
	}
	if f.Tag != tag {
		t.Errorf("Field.Tag = %q, want %q", f.Tag, tag)
	}
	if f.Type != typ {
		t.Errorf("Field.Type = %q, want %q", f.Type, typ)
	}
}
