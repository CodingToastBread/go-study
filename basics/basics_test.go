package main

import (
	"errors"
	"testing"
)

// 12. 테이블 테스트 — Go 테스트의 기본 형태.
//   - 파일 이름은 _test.go, 함수 이름은 TestXxx(t *testing.T).
//   - 케이스를 데이터로 나열하고 루프 하나로 돌립니다.
//   - t.Run 으로 케이스마다 이름을 붙이면 `go test -run TestFindWatch/empty_id` 처럼 하나만 돌릴 수 있습니다.
//   - 실행: go test ./...      자세히: go test -v ./...      레이스 검사: go test -race ./...
func TestFindWatch(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error // nil 이면 성공을 기대
	}{
		{name: "found", id: "w_1", wantErr: nil},
		{name: "not found", id: "w_404", wantErr: ErrNotFound},
		{name: "empty id", id: "", wantErr: &ValidationError{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := findWatch(tc.id)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if w.ID != tc.id {
					t.Errorf("got ID %q, want %q", w.ID, tc.id)
				}
				return
			}

			var ve *ValidationError
			switch {
			case errors.Is(tc.wantErr, ErrNotFound):
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("got %v, want ErrNotFound", err)
				}
			case errors.As(tc.wantErr, &ve):
				if !errors.As(err, &ve) {
					t.Errorf("got %v, want *ValidationError", err)
				}
			}
		})
	}
}

func TestWatchRecord(t *testing.T) {
	w := NewWatch("w_1", "https://example.com")
	if w.Record("100") {
		t.Error("first value should not count as a change")
	}
	if w.Record("100") {
		t.Error("same value should not be a change")
	}
	if !w.Record("90") {
		t.Error("different value should be a change")
	}
}

func TestSum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum = %d, want 6", got)
	}
}
