package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── 테스트 ─────────────────────────────────────────────────────────────────
// 파일 이름 *_test.go, 함수 이름 TestXxx(t *testing.T). 같은 패키지라 비공개 함수도 부를 수 있습니다.
//   go test ./05-stdlib            전부
//   go test -run TestHello -v      하나만, 자세히
//   go test -race ./...            레이스 검사
//   go test -bench . -benchmem     벤치마크

// 테이블 테스트: 케이스를 데이터로 나열하고 루프 하나로 돌립니다.
func TestHello(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantHello  string
	}{
		{"ok", "/hello/go", http.StatusOK, "go"},
		{"korean", "/hello/상언", http.StatusOK, "상언"},
		{"unknown route", "/nope", http.StatusNotFound, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + tc.path)
			if err != nil {
				t.Fatalf("request failed: %v", err) // Fatal: 이 서브테스트를 즉시 중단
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus) // Error: 기록하고 계속
			}
			if tc.wantHello == "" {
				return
			}
			var got map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got["hello"] != tc.wantHello {
				t.Errorf("hello = %q, want %q", got["hello"], tc.wantHello)
			}
		})
	}
}

// httptest.NewRecorder: 서버를 띄우지 않고 핸들러만 직접 호출하는 더 가벼운 방식.
func TestEchoWithRecorder(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", nil)
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

// 벤치마크: b.N 번 반복. 결과는 ns/op 와 allocs/op.
func BenchmarkPriceRegex(b *testing.B) {
	s := "가격 ₩1,890,000 (지난주 2,100,000)"
	for i := 0; i < b.N; i++ {
		priceRE.FindAllString(s, -1)
	}
}

// t.Helper 를 부른 함수는 실패 위치가 호출한 쪽 줄로 찍힙니다.
func mustParse(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("bad json %q: %v", s, err)
	}
	return m
}

func TestHelper(t *testing.T) {
	m := mustParse(t, `{"a":1}`)
	if m["a"] != float64(1) {
		t.Errorf("a = %v", m["a"])
	}
}
