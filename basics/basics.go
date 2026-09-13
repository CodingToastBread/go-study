// basics.go — Go 기본기 한 장.
//
// 실행:  go run basics.go
// 사용법: 위에서 아래로 읽으면서 출력과 대조하세요. 궁금한 블록은 숫자를
// 바꾸거나 한 줄을 지우고 다시 돌려보세요. 각 블록은 독립적입니다.
//
// 순서: 변수·타입, 슬라이스·맵, 구조체·메서드, 에러, 인터페이스, 클로저·defer,
// 제네릭, goroutine·채널, select, WaitGroup, context, 그리고 테스트(basics_test.go).
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	section("1. 변수와 타입")
	basicTypes()

	section("2. 슬라이스와 맵")
	slicesAndMaps()

	section("3. 구조체와 메서드")
	structsAndMethods()

	section("4. 에러는 값이다")
	errorsAreValues()

	section("5. 인터페이스 — 암묵적 구현, 작게")
	interfaces()

	section("6. 클로저와 defer")
	closuresAndDefer()

	section("7. 제네릭")
	generics()

	section("8. goroutine과 채널")
	goroutinesAndChannels()

	section("9. select — 여러 채널을 동시에 듣기")
	selectLoop()

	section("10. sync.WaitGroup — 전부 끝날 때까지")
	waitGroup()

	section("11. context — 취소와 타임아웃")
	contexts()

	section("12. 테스트는 basics_test.go 를 보세요 → go test ./...")
}

func section(title string) {
	fmt.Printf("\n── %s ──\n", title)
}

// ───────────────────────────────────────────────────────────────────────────
// 1. 변수와 타입
//   - := 는 선언+초기화(함수 안에서만). var 는 어디서나.
//   - 선언했는데 안 쓰면 컴파일 에러. 그게 Go의 성격입니다.
//   - 모든 타입은 제로값이 있어서 nil 체크 없이 바로 씁니다.
func basicTypes() {
	name := "pricewatch" // string
	port := 8080         // int
	ratio := 0.05        // float64
	enabled := true      // bool

	var count int      // 제로값 0
	var label string   // 제로값 ""
	var when time.Time // 제로값 0001-01-01
	var maybe *int     // 포인터 제로값 nil

	fmt.Println(name, port, ratio, enabled)
	fmt.Println("zero values:", count, label == "", when.IsZero(), maybe == nil)

	// 타입 정의: 기존 타입에 이름을 붙이면 메서드를 달 수 있고, 섞어 쓰면 컴파일 에러.
	type WatchID string
	id := WatchID("w_123")
	fmt.Printf("%v has type %T\n", id, id)

	// 상수와 iota: 열거형 대신 씁니다.
	type Status int
	const (
		Pending Status = iota // 0
		Running               // 1
		Failed                // 2
	)
	fmt.Println("statuses:", Pending, Running, Failed)

	// 여러 값 반환은 언어 기본 기능. 에러 처리의 토대입니다.
	q, r := divmod(17, 5)
	fmt.Println("17 / 5 =", q, "remainder", r)
}

func divmod(a, b int) (int, int) { return a / b, a % b }

// ───────────────────────────────────────────────────────────────────────────
// 2. 슬라이스와 맵
//   - 슬라이스는 "배열을 보는 창". append 하면 새 슬라이스가 돌아옵니다(재대입 필수).
//   - 맵은 참조 타입. nil 맵에 쓰면 panic → make 로 만듭니다.
//   - 순회 순서는 보장되지 않습니다. 순서가 필요하면 키를 정렬합니다.
func slicesAndMaps() {
	prices := []int{1890000, 1690000}
	prices = append(prices, 1590000) // 재대입! append 결과를 버리면 사라집니다.
	fmt.Println("prices:", prices, "len:", len(prices))

	last := prices[len(prices)-1]
	firstTwo := prices[:2] // 같은 배열을 공유하는 창. 복사가 아닙니다.
	fmt.Println("last:", last, "firstTwo:", firstTwo)

	for i, p := range prices { // 인덱스와 값. 값만 필요하면 for _, p := range
		fmt.Printf("  [%d] %d\n", i, p)
	}

	byKind := map[string]int{"slack": 2, "discord": 1}
	byKind["telegram"]++ // 없는 키를 읽으면 제로값(0)이 나와서 ++ 가 그냥 됩니다.

	v, ok := byKind["email"] // "있는지"는 두 번째 반환값으로 확인합니다.
	fmt.Println("email:", v, "exists:", ok)

	keys := make([]string, 0, len(byKind))
	for k := range byKind {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("kinds sorted:", keys)
}

// ───────────────────────────────────────────────────────────────────────────
// 3. 구조체와 메서드
//   - 클래스가 없습니다. 데이터(struct)와 동작(메서드)을 따로 정의하고 리시버로 묶습니다.
//   - 포인터 리시버(*Watch)는 값을 바꿀 때, 값 리시버(Watch)는 읽기만 할 때.
//     한 타입 안에서는 하나로 통일하는 게 관례입니다.
//   - 생성자는 그냥 NewXxx 함수. 필드는 대문자로 시작하면 패키지 밖에 공개됩니다.
type Watch struct {
	ID        string
	URL       string
	Interval  time.Duration
	LastValue string
	checks    int // 소문자 → 패키지 밖에서 안 보임
}

func NewWatch(id, url string) *Watch {
	return &Watch{ID: id, URL: url, Interval: 10 * time.Minute}
}

func (w *Watch) Record(value string) (changed bool) {
	w.checks++
	changed = w.LastValue != "" && w.LastValue != value
	w.LastValue = value
	return changed
}

func (w Watch) String() string { // fmt 가 String() 을 알아서 씁니다(Stringer 인터페이스).
	return fmt.Sprintf("Watch(%s every %s, checked %d times)", w.ID, w.Interval, w.checks)
}

func structsAndMethods() {
	w := NewWatch("w_1", "https://example.com/item")
	fmt.Println("changed?", w.Record("1890000"))
	fmt.Println("changed?", w.Record("1890000"))
	fmt.Println("changed?", w.Record("1690000"))
	fmt.Println(w)

	// 구조체 임베딩: 상속이 아니라 "필드와 메서드를 끌어올리기"입니다.
	type Timestamped struct{ CreatedAt time.Time }
	type Check struct {
		Timestamped
		Value string
	}
	c := Check{Timestamped{time.Now()}, "1690000"}
	fmt.Println("embedded field promoted:", !c.CreatedAt.IsZero())
}

// ───────────────────────────────────────────────────────────────────────────
// 4. 에러는 값이다
//   - 예외가 없습니다. 실패할 수 있는 함수는 마지막 반환값으로 error 를 돌려줍니다.
//   - 호출한 쪽은 즉시 `if err != nil` 로 결정합니다: 처리하거나, 감싸서 올리거나.
//   - 감쌀 때는 fmt.Errorf("문맥: %w", err). 위에서는 errors.Is / errors.As 로 분기.
//   - 패키지가 공개한 에러 변수(센티널)로 "없음" 같은 정상적인 실패를 약속합니다.
var ErrNotFound = errors.New("watch: not found")

type ValidationError struct{ Field, Reason string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

func findWatch(id string) (*Watch, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Reason: "must not be empty"}
	}
	if id != "w_1" {
		return nil, fmt.Errorf("find watch %q: %w", id, ErrNotFound) // 문맥을 붙여 감싸기
	}
	return NewWatch(id, "https://example.com"), nil
}

func errorsAreValues() {
	for _, id := range []string{"w_1", "w_404", ""} {
		w, err := findWatch(id)
		switch {
		case err == nil:
			fmt.Println("found:", w.ID)
		case errors.Is(err, ErrNotFound): // 감싸져 있어도 찾아냅니다
			fmt.Println("404 →", err)
		default:
			var ve *ValidationError
			if errors.As(err, &ve) { // 타입으로 꺼내기
				fmt.Println("400 → field:", ve.Field, "reason:", ve.Reason)
			}
		}
	}
	// panic 은 "프로그래머 실수(버그)"에만. 정상적인 실패에는 쓰지 않습니다.
}

// ───────────────────────────────────────────────────────────────────────────
// 5. 인터페이스 — 암묵적 구현, 작게
//   - implements 키워드가 없습니다. 메서드 집합이 맞으면 그냥 구현한 겁니다.
//   - 인터페이스는 "쓰는 쪽"이 필요한 만큼만 작게 정의합니다(1~3개 메서드).
//   - 그래서 테스트용 가짜(fake)를 몇 줄로 만들 수 있습니다.
type Notifier interface {
	Send(msg string) error
}

type slackNotifier struct{ webhook string }

func (s slackNotifier) Send(msg string) error {
	fmt.Println("  [slack]", msg)
	return nil
}

type stdoutNotifier struct{}

func (stdoutNotifier) Send(msg string) error {
	fmt.Println("  [stdout]", msg)
	return nil
}

// 쓰는 쪽은 구체 타입을 모릅니다. Notifier 만 요구합니다.
func alert(n Notifier, w *Watch, old, cur string) error {
	return n.Send(fmt.Sprintf("%s: %s → %s", w.ID, old, cur))
}

func interfaces() {
	w := NewWatch("w_1", "https://example.com")
	notifiers := map[string]Notifier{ // 종류별로 골라 끼우기
		"slack":  slackNotifier{webhook: "https://hooks.slack.com/..."},
		"stdout": stdoutNotifier{},
	}
	for _, kind := range []string{"slack", "stdout"} {
		_ = alert(notifiers[kind], w, "1890000", "1690000")
	}

	// 타입 단언과 타입 스위치: 인터페이스 값 안의 구체 타입을 확인합니다.
	var n Notifier = stdoutNotifier{}
	if _, ok := n.(stdoutNotifier); ok {
		fmt.Println("  n is stdoutNotifier")
	}
	switch v := n.(type) {
	case slackNotifier:
		fmt.Println("  slack with", v.webhook)
	case stdoutNotifier:
		fmt.Println("  stdout")
	}
	// 빈 인터페이스 any(= interface{}) 는 "무엇이든". 꼭 필요할 때만 씁니다.
}

// ───────────────────────────────────────────────────────────────────────────
// 6. 클로저와 defer
//   - 함수는 값입니다. 변수에 담고, 인자로 넘기고, 바깥 변수를 붙잡습니다(클로저).
//   - defer 는 "이 함수를 나갈 때 실행". 파일 닫기, 락 해제, 트랜잭션 롤백에 씁니다.
//   - 여러 defer 는 역순(LIFO)으로 실행됩니다.
func closuresAndDefer() {
	counter := func() func() int {
		n := 0
		return func() int { n++; return n } // n 을 붙잡고 있습니다
	}()
	fmt.Println("counter:", counter(), counter(), counter())

	func() {
		defer fmt.Println("  defer 1 (마지막에 실행)")
		defer fmt.Println("  defer 2 (그 전에 실행)")
		fmt.Println("  함수 본문")
	}()

	// 흔한 실제 용례: 열었으면 닫는다. 열기와 닫기를 나란히 씁니다.
	f, err := os.CreateTemp("", "basics-*.txt")
	if err != nil {
		fmt.Println("temp file:", err)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_, _ = f.WriteString("hello")
	fmt.Println("  wrote temp file:", strings.HasSuffix(f.Name(), ".txt"))
}

// ───────────────────────────────────────────────────────────────────────────
// 7. 제네릭 (Go 1.18+)
//   - 타입 파라미터 [T any]. 제약(constraint)으로 허용 타입을 좁힙니다.
//   - 슬라이스/맵 유틸에 주로 씁니다. 과하게 쓰면 읽기 어려워집니다.
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

type Number interface{ ~int | ~int64 | ~float64 }

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func generics() {
	prices := []int{1890000, 1690000, 1590000}
	labels := Map(prices, func(p int) string { return fmt.Sprintf("₩%d", p) })
	fmt.Println("labels:", labels)
	fmt.Println("sum:", Sum(prices), "avg:", Sum([]float64{1.5, 2.5})/2)
}

// ───────────────────────────────────────────────────────────────────────────
// 8. goroutine과 채널
//   - `go f()` 로 동시 실행을 시작합니다. 수천 개를 띄워도 가볍습니다.
//   - 채널은 goroutine 사이의 타입 있는 파이프. 보내면(ch <- v) 받을 때까지(<-ch) 기다립니다.
//   - 채널을 닫는(close) 쪽은 항상 "보내는 쪽". 받는 쪽은 range 로 닫힐 때까지 받습니다.
func goroutinesAndChannels() {
	urls := []string{"https://a.example", "https://b.example", "https://c.example"}

	type result struct {
		url  string
		took time.Duration
	}
	results := make(chan result) // 버퍼 없는 채널

	for _, u := range urls {
		go func() { // 1.22+ 부터 루프 변수 u 가 반복마다 새로 만들어집니다
			start := time.Now()
			time.Sleep(time.Duration(10+len(u)) * time.Millisecond) // 가짜 네트워크 지연
			results <- result{u, time.Since(start)}
		}()
	}

	for range urls { // 보낸 개수만큼 받습니다
		r := <-results
		fmt.Printf("  %s in %v\n", r.url, r.took.Round(time.Millisecond))
	}

	// 파이프라인: 생산자가 닫고, 소비자는 range.
	nums := make(chan int, 3) // 버퍼 3: 받는 쪽이 없어도 3개까지는 보내고 지나갑니다
	go func() {
		defer close(nums) // 보내는 쪽이 닫는다
		for i := 1; i <= 3; i++ {
			nums <- i * i
		}
	}()
	for n := range nums {
		fmt.Print("  squared:", n, " ")
	}
	fmt.Println()
}

// ───────────────────────────────────────────────────────────────────────────
// 9. select — 여러 채널을 동시에 듣기
//   - 준비된 case 하나를 실행합니다. 여러 개가 준비되면 무작위.
//   - "주기 작업 + 종료 신호"를 한 루프에서 듣는 패턴이 스케줄러의 뼈대입니다.
func selectLoop() {
	ticker := time.NewTicker(15 * time.Millisecond)
	defer ticker.Stop()
	done := time.After(60 * time.Millisecond)

	ticks := 0
	for {
		select {
		case <-ticker.C:
			ticks++
			fmt.Println("  tick", ticks)
		case <-done:
			fmt.Println("  done after", ticks, "ticks")
			return
		}
	}
}

// ───────────────────────────────────────────────────────────────────────────
// 10. sync.WaitGroup — 전부 끝날 때까지
//   - Add(n) 로 기다릴 개수를 등록, 각 goroutine 이 끝날 때 Done(), Wait() 로 대기.
//   - 공유 변수를 여러 goroutine 이 쓰면 sync.Mutex 로 보호합니다.
//     `go test -race` 가 보호 안 된 곳을 잡아줍니다.
func waitGroup() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			total += i
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("  total:", total)
	// Go 1.25+ 에는 wg.Go(func(){...}) 도 있습니다: Add/Done 을 대신 해줍니다.
}

// ───────────────────────────────────────────────────────────────────────────
// 11. context — 취소와 타임아웃
//   - 첫 번째 인자로 ctx 를 넘기는 게 관례. 구조체에 저장하지 않습니다.
//   - WithTimeout / WithCancel 로 자식 ctx 를 만들고, 반드시 cancel() 을 defer 합니다.
//   - 오래 걸리는 작업은 ctx.Done() 을 듣고 있다가 신호가 오면 멈춥니다.
func slowFetch(ctx context.Context, url string, takes time.Duration) error {
	select {
	case <-time.After(takes):
		fmt.Println("  fetched", url)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("fetch %s: %w", url, ctx.Err()) // context.DeadlineExceeded 등
	}
}

func contexts() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_ = slowFetch(ctx, "https://fast.example", 10*time.Millisecond)
	err := slowFetch(ctx, "https://slow.example", 100*time.Millisecond)
	fmt.Println("  slow:", err)
	fmt.Println("  is timeout?", errors.Is(err, context.DeadlineExceeded))
}
