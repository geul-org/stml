package validator

import "fmt"

// ValidationError represents a single validation failure.
type ValidationError struct {
	File    string // source HTML filename
	Attr    string // attribute context (e.g. `data-fetch="Login"`)
	Message string // human-readable error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("ERROR: %s — %s: %s", e.File, e.Attr, e.Message)
}

func errOpNotFound(file, attr, op string) ValidationError {
	return ValidationError{file, attr, fmt.Sprintf("OpenAPI에 %q operationId가 없습니다", op)}
}

func errWrongMethod(file, attr, op, got, want string) ValidationError {
	return ValidationError{file, attr, fmt.Sprintf("%q은 %s 메서드입니다 (%s이어야 함)", op, got, want)}
}

func errParamNotFound(file, op, param string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-param-%s", param), fmt.Sprintf("%q의 parameters에 %q가 없습니다", op, param)}
}

func errFieldNotFound(file, op, field string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-field=%q", field), fmt.Sprintf("%q의 request schema에 %q 필드가 없습니다", op, field)}
}

func errBindNotFound(file, op, field string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-bind=%q", field), fmt.Sprintf("%q의 response schema에도, custom.ts에도 %q가 없습니다", op, field)}
}

func errEachNotArray(file, op, field string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-each=%q", field), fmt.Sprintf("%q의 response에서 %q는 배열이 아닙니다", op, field)}
}

func errEachNotFound(file, op, field string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-each=%q", field), fmt.Sprintf("%q의 response에 %q 필드가 없습니다", op, field)}
}

func errComponentNotFound(file, comp, path string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-component=%q", comp), fmt.Sprintf("%s 파일이 없습니다", path)}
}

func errPaginateNoExt(file, op string) ValidationError {
	return ValidationError{file, "data-paginate", fmt.Sprintf("%q에 x-pagination이 선언되지 않았습니다", op)}
}

func errSortNotAllowed(file, op, col string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-sort=%q", col), fmt.Sprintf("%q의 x-sort.allowed에 %q가 없습니다", op, col)}
}

func errFilterNotAllowed(file, op, col string) ValidationError {
	return ValidationError{file, fmt.Sprintf("data-filter=%q", col), fmt.Sprintf("%q의 x-filter.allowed에 %q가 없습니다", op, col)}
}
