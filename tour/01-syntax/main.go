// 01-syntax — 값, 변수, 상수, 제어문, 함수, 클로저, 포인터.
//
// 실행: go run ./01-syntax
// 다섯 묶음 중 첫 번째. 주석을 따라 위에서 아래로 읽고, 숫자를 바꿔 다시 돌려보세요.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

func main() {
	values()
	variablesAndConstants()
	loops()
	branches()
	functions()
	closures()
	pointers()
}

// ── 값 ─────────────────────────────────────────────────────────────────────
// 문자열은 + 로 잇고, 숫자는 정수와 실수가 섞이지 않습니다(명시적 변환 필요).
func values() {
	fmt.Println("go" + "lang")
	fmt.Println("7/2 =", 7/2, "  7.0/2 =", 7.0/2, "  7%2 =", 7%2)
	fmt.Println(true && !false, 1 < 2)

	var n int = 7
	var f float64 = float64(n) / 2 // int → float64 는 반드시 명시적으로
	fmt.Println("converted:", f)
	fmt.Println()
}

// ── 변수와 상수 ─────────────────────────────────────────────────────────────
// var 는 어디서나, := 는 함수 안에서만. 초기값이 없으면 제로값.
// 상수는 타입이 없을 수도 있고(untyped), 필요한 곳에서 타입이 정해집니다.
func variablesAndConstants() {
	var a = "initial"
	var b, c int = 1, 2
	var d bool   // 제로값 false
	e := "short" // var e string = "short" 와 같음
	fmt.Println(a, b, c, d, e)

	const big = 1 << 40                 // untyped 상수. int 범위를 넘어도 상수 자체는 OK.
	fmt.Println(big/1e9, math.Sin(big)) // 쓰이는 자리에서 float64 가 됨

	type Level int
	const (
		Low  Level = iota // 0
		Mid               // 1
		High              // 2
	)
	fmt.Println("levels:", Low, Mid, High)
	fmt.Println()
}

// ── 반복 ───────────────────────────────────────────────────────────────────
// 반복문은 for 하나뿐입니다. while, do-while 도 전부 for 로 씁니다.
func loops() {
	for i := 0; i < 3; i++ { // 고전적
		fmt.Print(i, " ")
	}
	fmt.Println()

	n := 3
	for n > 0 { // while
		n--
	}
	fmt.Println("n after while:", n)

	for i := range 3 { // Go 1.22+: 0,1,2
		fmt.Print(i, " ")
	}
	fmt.Println()

	for { // 무한 루프 + break
		n++
		if n == 2 {
			break
		}
	}

	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			continue // 짝수 건너뛰기
		}
		fmt.Print("odd:", i, " ")
	}
	fmt.Println()
	fmt.Println()
}

// ── 분기 ───────────────────────────────────────────────────────────────────
// if 에 괄호 없음, 중괄호 필수. 조건 앞에 짧은 문장을 둘 수 있습니다.
// switch 는 break 가 자동이고, 조건 없는 switch 는 if-else 체인을 대신합니다.
func branches() {
	if n := 9; n%2 == 0 { // n 은 if 블록 안에서만 살아 있음
		fmt.Println(n, "is even")
	} else if n < 10 {
		fmt.Println(n, "is odd and small")
	} else {
		fmt.Println(n, "is odd")
	}

	switch day := "sat"; day {
	case "sat", "sun": // 여러 값
		fmt.Println("weekend")
	default:
		fmt.Println("weekday")
	}

	t := 15
	switch { // 조건 없는 switch
	case t < 12:
		fmt.Println("morning")
	case t < 18:
		fmt.Println("afternoon")
	default:
		fmt.Println("evening")
	}

	var x any = 3.14
	switch v := x.(type) { // 타입 switch
	case int:
		fmt.Println("int", v)
	case float64:
		fmt.Println("float64", v)
	default:
		fmt.Printf("other %T\n", v)
	}
	fmt.Println()
}

// ── 함수 ───────────────────────────────────────────────────────────────────
// 타입이 이름 뒤에 옵니다. 여러 값을 반환할 수 있고, 가변 인자는 ...T.
func add(a, b int) int { return a + b }

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide %v by zero", a)
	}
	return a / b, nil
}

func sum(nums ...int) (total int) { // 이름 있는 반환값: 선언과 동시에 제로값
	for _, n := range nums {
		total += n
	}
	return // bare return: total 을 그대로 반환
}

func functions() {
	fmt.Println("add:", add(1, 2))

	if q, err := divide(1, 0); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("quotient:", q)
	}

	fmt.Println("sum:", sum(1, 2, 3))
	nums := []int{4, 5, 6}
	fmt.Println("sum of slice:", sum(nums...)) // 슬라이스를 펼쳐서 전달
	fmt.Println()
}

// ── 클로저 ─────────────────────────────────────────────────────────────────
// 함수는 값입니다. 바깥 변수를 붙잡은 함수를 돌려주면 그 변수는 계속 살아 있습니다.
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func closures() {
	next := counter()
	fmt.Println(next(), next(), next()) // 1 2 3

	other := counter() // 독립된 n
	fmt.Println(other())

	// 함수를 인자로 넘기기
	apply := func(xs []string, f func(string) string) []string {
		out := make([]string, len(xs))
		for i, x := range xs {
			out[i] = f(x)
		}
		return out
	}
	fmt.Println(apply([]string{"a", "b"}, strings.ToUpper))
	fmt.Println()
}

// ── 포인터 ─────────────────────────────────────────────────────────────────
// & 로 주소를 얻고 * 로 값을 따라갑니다. 포인터 연산은 없습니다.
// 함수 인자는 항상 복사되므로, 바꾸고 싶으면 포인터를 넘깁니다.
func zeroByValue(n int)    { n = 0 }
func zeroByPointer(n *int) { *n = 0 }

func pointers() {
	n := 42
	zeroByValue(n)
	fmt.Println("after by-value:", n)
	zeroByPointer(&n)
	fmt.Println("after by-pointer:", n)

	p := &n
	fmt.Println("pointer holds address:", p != nil, " value:", *p)

	// 흔한 실전 모양: 구조체는 포인터로 넘겨 복사를 피하고 수정도 합니다.
	type Config struct{ Port int }
	setPort := func(c *Config, port int) { c.Port = port }
	cfg := Config{Port: 8080}
	setPort(&cfg, 9090)
	fmt.Println("port:", cfg.Port)

	fmt.Fprintln(os.Stderr, "(stderr 에도 쓸 수 있습니다)")
}
