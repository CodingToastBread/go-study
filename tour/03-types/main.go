// 03-types — 구조체, 메서드, 임베딩, 인터페이스, 에러, panic/recover, defer.
//
// 실행: go run ./03-types
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	structs()
	methods()
	embedding()
	interfaces()
	errorsBasics()
	errorsWrapping()
	deferAndPanic()
}

// ── 구조체 ─────────────────────────────────────────────────────────────────
// 필드의 묶음. 클래스가 아닙니다. 값 타입이라 대입하면 복사됩니다.
type Watch struct {
	ID       string
	URL      string
	Interval time.Duration
	tags     []string // 소문자: 패키지 밖에서 안 보임
}

func structs() {
	w := Watch{ID: "w_1", URL: "https://a.example"} // 나머지 필드는 제로값
	w2 := w                                         // 복사
	w2.ID = "w_2"
	fmt.Println(w.ID, w2.ID)

	p := &Watch{ID: "w_3"} // 포인터. 필드 접근은 자동 역참조(p.ID, (*p).ID 아님)
	p.Interval = time.Minute
	fmt.Printf("%+v\n", *p) // %+v 는 필드 이름까지

	anon := struct{ X, Y int }{1, 2} // 익명 구조체: 한 번만 쓸 때
	fmt.Println(anon)
	fmt.Println()
}

// ── 메서드 ─────────────────────────────────────────────────────────────────
// 리시버가 붙은 함수. 포인터 리시버는 수정 가능, 값 리시버는 복사본에서 동작.
// 한 타입 안에서는 하나로 통일하는 게 관례입니다(대개 포인터).
func (w *Watch) AddTag(t string) { w.tags = append(w.tags, t) }
func (w Watch) Tags() string     { return strings.Join(w.tags, ",") }

// 기본 타입에도 이름을 붙이면 메서드를 달 수 있습니다.
type Celsius float64

func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func methods() {
	w := Watch{ID: "w_1"}
	w.AddTag("laptop") // Go 가 (&w).AddTag 로 바꿔줍니다
	w.AddTag("sale")
	fmt.Println(w.Tags())

	temp := Celsius(36.6)
	fmt.Println(temp) // fmt 가 String() 을 알아서 씁니다
	fmt.Println()
}

// ── 임베딩 ─────────────────────────────────────────────────────────────────
// 상속이 아닙니다. 필드와 메서드를 "끌어올리는" 것입니다.
type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Timestamps) Touch() { t.UpdatedAt = time.Now() }

type Channel struct {
	Timestamps // 이름 없는 필드 = 임베딩
	Name       string
}

func embedding() {
	c := Channel{Name: "slack"}
	c.Touch()                                    // Timestamps 의 메서드가 끌어올려짐
	fmt.Println(c.UpdatedAt.IsZero())            // false. c.Timestamps.UpdatedAt 와 같음
	fmt.Println(c.Timestamps.CreatedAt.IsZero()) // 명시적으로도 접근 가능
	fmt.Println()
}

// ── 인터페이스 ─────────────────────────────────────────────────────────────
// 메서드 집합. implements 키워드가 없고, 메서드가 맞으면 자동으로 구현입니다.
// 작게(1~3개 메서드), 그리고 "쓰는 쪽"이 정의합니다.
type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }

func (r Rect) Area() float64   { return r.W * r.H }
func (c Circle) Area() float64 { return 3.14159 * c.R * c.R }

func totalArea(shapes ...Shape) float64 {
	sum := 0.0
	for _, s := range shapes {
		sum += s.Area()
	}
	return sum
}

func interfaces() {
	fmt.Printf("%.2f\n", totalArea(Rect{2, 3}, Circle{1}))

	// 타입 단언: 인터페이스 값 안의 구체 타입 꺼내기
	var s Shape = Rect{2, 3}
	if r, ok := s.(Rect); ok {
		fmt.Println("it's a rect with W =", r.W)
	}

	// 인터페이스 조합
	type Named interface{ Name() string }
	type NamedShape interface {
		Shape
		Named
	}
	var _ NamedShape // 선언만. 조합이 된다는 것을 보여주는 용도.

	// 자주 보는 표준 인터페이스: fmt.Stringer, error, io.Reader, io.Writer, sort.Interface
	var sb strings.Builder // io.Writer 를 구현
	fmt.Fprintf(&sb, "written via io.Writer")
	fmt.Println(sb.String())
	fmt.Println()
}

// ── 에러 ───────────────────────────────────────────────────────────────────
// error 는 Error() string 하나짜리 인터페이스입니다. 반환값으로 흐릅니다.
var ErrNotFound = errors.New("not found") // 센티널: 패키지가 공개하는 에러 값

type ValidationError struct{ Field, Reason string } // 정보가 필요하면 타입으로

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

func find(id string) (Watch, error) {
	switch id {
	case "":
		return Watch{}, &ValidationError{"id", "empty"}
	case "w_1":
		return Watch{ID: id}, nil
	default:
		return Watch{}, ErrNotFound
	}
}

func errorsBasics() {
	for _, id := range []string{"w_1", "w_9", ""} {
		w, err := find(id)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println("found:", w.ID)
	}
	fmt.Println()
}

// ── 에러 감싸기 ─────────────────────────────────────────────────────────────
// %w 로 감싸 올리면 문맥이 쌓이고, errors.Is / As 는 감싸진 것을 뚫고 찾습니다.
func loadWatch(id string) error {
	_, err := find(id)
	if err != nil {
		return fmt.Errorf("load watch %q: %w", id, err)
	}
	return nil
}

func errorsWrapping() {
	err := loadWatch("w_9")
	fmt.Println(err)                         // load watch "w_9": not found
	fmt.Println(errors.Is(err, ErrNotFound)) // true — 감싸져 있어도 찾음
	fmt.Println(errors.Unwrap(err) == ErrNotFound)

	err = loadWatch("")
	var ve *ValidationError
	if errors.As(err, &ve) { // 타입으로 꺼내기
		fmt.Println("validation on field:", ve.Field)
	}

	joined := errors.Join(ErrNotFound, errors.New("also this")) // 여러 에러를 하나로
	fmt.Println(errors.Is(joined, ErrNotFound), "|", strings.ReplaceAll(joined.Error(), "\n", " / "))
	fmt.Println()
}

// ── defer, panic, recover ──────────────────────────────────────────────────
// defer: 함수를 나갈 때 실행(역순). 파일 닫기, 락 해제, 트랜잭션 롤백.
// panic: 프로그래머 실수(버그)용. 정상적인 실패에는 error 를 씁니다.
// recover: defer 안에서만 panic 을 잡습니다. 서버가 통째로 죽는 걸 막을 때.
func risky(n int) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	arr := []int{1, 2, 3}
	return fmt.Sprint(arr[n]), nil // n 이 범위 밖이면 panic
}

func deferAndPanic() {
	f, err := os.CreateTemp("", "tour-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(f.Name()) // 나중에 실행
	defer f.Close()           // 먼저 실행 (역순)
	fmt.Println("temp file open:", f.Name() != "")

	fmt.Println(risky(1))
	fmt.Println(risky(10))

	for i := 0; i < 3; i++ {
		defer fmt.Print(i, " ") // 3개가 쌓였다가 함수 끝에 2 1 0 순으로
	}
	fmt.Print("deferred prints: ")
}
