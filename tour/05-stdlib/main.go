// 05-stdlib — 실무에서 매일 쓰는 표준 라이브러리: fmt 포맷, 시간, JSON, 정규식,
// 파일, 환경변수/인자, HTTP 클라이언트와 서버(httptest), 그리고 테스트.
//
// 실행: go run ./05-stdlib
// 테스트: go test ./05-stdlib
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	formatting()
	timeBasics()
	jsonBasics()
	regexBasics()
	files()
	envAndArgs()
	httpClientAndServer()
}

// ── fmt ────────────────────────────────────────────────────────────────────
// %v 는 뭐든, %+v 는 필드 이름까지, %#v 는 Go 문법으로, %T 는 타입. %q 는 따옴표 문자열.
func formatting() {
	type P struct {
		X, Y int
	}
	p := P{1, 2}
	fmt.Printf("%v %+v %#v %T\n", p, p, p, p)
	fmt.Printf("%d %5d %-5d| %05d\n", 42, 42, 42, 42)
	fmt.Printf("%.2f %8.3f %e\n", 3.14159, 3.14159, 123456.789)
	fmt.Printf("%s %q %10s|%-10s|\n", "go", "go", "right", "left")
	fmt.Printf("%t %c %x %b %p\n", true, 'A', 255, 5, &p)

	s := fmt.Sprintf("id=%d", 7) // 문자열로
	fmt.Fprintln(os.Stderr, s)   // 다른 곳으로

	n, err := strconv.Atoi("123") // 문자열 ↔ 숫자는 strconv
	f, _ := strconv.ParseFloat("1.5", 64)
	b, _ := strconv.ParseBool("true")
	fmt.Println(n, err, f, b, strconv.Itoa(99), strconv.Quote("hi"))
	fmt.Println()
}

// ── time ───────────────────────────────────────────────────────────────────
// time.Time 은 값이고, Duration 은 int64 나노초입니다. 포맷은 기준 시각 "2006-01-02 15:04:05".
func timeBasics() {
	now := time.Now()
	fmt.Println(now.Format(time.RFC3339), now.Format("2006-01-02 15:04"))

	t, err := time.Parse("2006-01-02", "2026-12-31")
	fmt.Println(t.Weekday(), err, t.Sub(now).Round(time.Hour) > 0)

	d := 90 * time.Minute
	fmt.Println(d, d.Hours(), d.Round(time.Hour))

	deadline := now.Add(10 * time.Minute)
	fmt.Println("due?", !now.Before(deadline), " until:", deadline.Sub(now))

	seoul, _ := time.LoadLocation("Asia/Seoul")
	fmt.Println(now.In(seoul).Format("15:04 MST"))
	fmt.Println()
}

// ── JSON ───────────────────────────────────────────────────────────────────
// 구조체 태그로 키 이름을 정합니다. 대문자 필드만 직렬화됩니다.
// omitempty 는 제로값을 생략, 포인터 필드는 "없음"과 "제로값"을 구분합니다.
type WatchDTO struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Threshold float64  `json:"threshold_pct,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Enabled   *bool    `json:"enabled,omitempty"` // nil 이면 키 자체가 없음
	secret    string   // 소문자: 직렬화 안 됨
}

func jsonBasics() {
	yes := true
	w := WatchDTO{ID: "w_1", URL: "https://a.example", Tags: []string{"x"}, Enabled: &yes, secret: "hidden"}
	out, _ := json.Marshal(w)
	fmt.Println(string(out))

	pretty, _ := json.MarshalIndent(map[string]any{"items": []int{1, 2}, "n": 2}, "", "  ")
	fmt.Println(string(pretty))

	var back WatchDTO
	err := json.Unmarshal([]byte(`{"id":"w_2","url":"https://b.example","threshold_pct":5}`), &back)
	fmt.Printf("%+v %v enabled=%v\n", back, err, back.Enabled == nil)

	// 모양을 모를 때는 map[string]any 로. 숫자는 float64 로 들어옵니다.
	var anyJSON map[string]any
	_ = json.Unmarshal([]byte(`{"a":1,"b":[true,"x"]}`), &anyJSON)
	fmt.Printf("%v %T\n", anyJSON["a"], anyJSON["a"])

	// 스트림: 요청 본문처럼 io.Reader 에서 바로 읽기. 모르는 필드는 거부할 수 있습니다.
	dec := json.NewDecoder(strings.NewReader(`{"id":"w_3","typo":1}`))
	dec.DisallowUnknownFields()
	fmt.Println("unknown field:", dec.Decode(&back))
	fmt.Println()
}

// ── 정규식 ─────────────────────────────────────────────────────────────────
// 패키지 레벨에서 한 번만 컴파일합니다. 함수 안에서 매번 컴파일하면 느립니다.
var priceRE = regexp.MustCompile(`-?\d{1,3}(?:,\d{3})+(?:\.\d+)?|-?\d+(?:\.\d+)?`)

func regexBasics() {
	s := "가격 ₩1,890,000 (지난주 2,100,000)"
	fmt.Println(priceRE.MatchString(s))
	fmt.Println(priceRE.FindString(s))
	fmt.Println(priceRE.FindAllString(s, -1))
	fmt.Println(priceRE.ReplaceAllString(s, "N"))

	re := regexp.MustCompile(`(\w+)@(\w+)\.com`)
	m := re.FindStringSubmatch("mail me at kim@example.com")
	fmt.Println(m[1], m[2]) // 그룹
	fmt.Println()
}

// ── 파일 ───────────────────────────────────────────────────────────────────
// 작은 파일은 os.ReadFile / os.WriteFile 한 줄. 큰 파일은 os.Open + bufio.
// 열었으면 defer Close. 경로 조합은 filepath.Join.
func files() {
	dir, _ := os.MkdirTemp("", "tour")
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "notes.txt")
	_ = os.WriteFile(path, []byte("line 1\nline 2\n"), 0o644)

	data, err := os.ReadFile(path)
	fmt.Print(string(data), err, "\n")

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	fmt.Fprintln(f, "line 3")
	f.Close()

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Println(e.Name(), info.Size(), "bytes")
	}

	_, err = os.Stat(filepath.Join(dir, "missing"))
	fmt.Println("missing exists?", !os.IsNotExist(err))
	fmt.Println()
}

// ── 환경변수와 인자 ─────────────────────────────────────────────────────────
func envAndArgs() {
	os.Setenv("PRICEWATCH_PORT", "8080")
	port := os.Getenv("PRICEWATCH_PORT")
	_, set := os.LookupEnv("PRICEWATCH_NOPE") // 있는지 구분
	fmt.Println(port, set)

	fmt.Println("program:", filepath.Base(os.Args[0]), "args:", os.Args[1:])
	// 플래그는 flag 패키지: flag.String("name", "default", "help") → flag.Parse()
	fmt.Println()
}

// ── HTTP ───────────────────────────────────────────────────────────────────
// 서버: 핸들러는 func(w http.ResponseWriter, r *http.Request). ServeMux 가 라우터.
// 클라이언트: http.Client 하나를 재사용, Timeout 필수, resp.Body 는 반드시 Close.
// httptest 로 진짜 포트를 열어 둘을 한 프로세스 안에서 테스트합니다.
func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"hello": r.PathValue("name")})
	})
	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		w.Write(body)
	})
	return mux
}

func httpClientAndServer() {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(srv.URL + "/hello/go")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var got map[string]string
	json.NewDecoder(resp.Body).Decode(&got)
	fmt.Println(resp.StatusCode, got)

	resp2, err := client.Post(srv.URL+"/echo", "text/plain", strings.NewReader("ping"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp2.Body.Close()
	echoed, _ := io.ReadAll(resp2.Body)
	fmt.Println(resp2.StatusCode, string(echoed))

	resp3, err := client.Get(srv.URL + "/nope")
	if err != nil {
		fmt.Println(err)
		return
	}
	resp3.Body.Close()
	fmt.Println("unknown route:", resp3.StatusCode)
}
