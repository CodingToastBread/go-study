// 02-collections — 배열, 슬라이스, 맵, range, 문자열과 룬, 정렬, 제네릭.
//
// 실행: go run ./02-collections
package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

func main() {
	arrays()
	slicesBasics()
	sliceGotchas()
	mapsBasics()
	stringsAndRunes()
	sorting()
	generics()
}

// ── 배열 ───────────────────────────────────────────────────────────────────
// 길이가 타입의 일부입니다([3]int 와 [4]int 는 다른 타입). 값으로 복사됩니다.
// 실무에서는 거의 슬라이스를 씁니다. 배열은 "슬라이스 뒤에 있는 것"으로 이해하면 됩니다.
func arrays() {
	var a [3]int // [0 0 0]
	a[1] = 10
	b := a // 복사!
	b[1] = 99
	fmt.Println("array:", a, b, len(a))

	grid := [2][2]int{{1, 2}, {3, 4}}
	fmt.Println("2d:", grid)
	fmt.Println()
}

// ── 슬라이스 ───────────────────────────────────────────────────────────────
// 슬라이스 = (배열 포인터, 길이 len, 용량 cap). 배열을 보는 "창"입니다.
// append 는 용량이 모자라면 새 배열을 만들고 새 슬라이스를 돌려줍니다 → 재대입 필수.
func slicesBasics() {
	var s []string          // nil 슬라이스. len 0, 그대로 append 가능.
	s = append(s, "a", "b") // 재대입!
	s = append(s, []string{"c", "d"}...)
	fmt.Println(s, "len:", len(s), "cap:", cap(s))

	m := make([]int, 3, 10) // 길이 3, 용량 10
	fmt.Println(m, len(m), cap(m))

	fmt.Println("s[1:3] =", s[1:3], " s[:2] =", s[:2], " s[2:] =", s[2:])

	c := make([]string, len(s))
	n := copy(c, s) // 진짜 복사
	fmt.Println("copied", n, "items:", c)

	// slices 패키지(1.21+): 자주 쓰는 연산이 들어 있습니다.
	fmt.Println("contains b?", slices.Contains(s, "b"), " index of c:", slices.Index(s, "c"))
	fmt.Println()
}

// ── 슬라이스의 함정 ─────────────────────────────────────────────────────────
// 같은 배열을 공유하는 두 슬라이스는 한쪽을 바꾸면 다른 쪽도 바뀝니다.
func sliceGotchas() {
	base := []int{1, 2, 3, 4, 5}
	head := base[:2] // base 와 같은 배열을 봄
	head[0] = 100
	fmt.Println("shared backing array:", base[0]) // 100

	// append 가 용량 안에서 일어나면 원본을 덮어씁니다.
	head = append(head, 999)
	fmt.Println("append overwrote base[2]:", base[2]) // 999

	// 독립시키려면 복사하거나 [low:high:max] 로 용량을 잘라둡니다.
	safe := base[:2:2] // cap 2 → append 시 새 배열
	safe = append(safe, 7)
	fmt.Println("safe append left base alone:", base[2], safe)
	fmt.Println()
}

// ── 맵 ─────────────────────────────────────────────────────────────────────
// 참조 타입. make 나 리터럴로 만듭니다. nil 맵에 쓰면 panic.
// 순회 순서는 매번 달라질 수 있습니다. 순서가 필요하면 키를 정렬.
func mapsBasics() {
	ages := map[string]int{"kim": 30, "lee": 25}
	ages["park"] = 41
	delete(ages, "lee")

	age, ok := ages["lee"] // 없는 키 → 제로값 + false
	fmt.Println("lee:", age, ok, " len:", len(ages))

	ages["new"]++ // 없는 키도 제로값에서 바로 연산 가능
	fmt.Println(ages)

	keys := slices.Sorted(maps.Keys(ages)) // 1.23+: 이터레이터 → 정렬된 슬라이스
	fmt.Println("sorted keys:", keys)

	// 값이 구조체인 맵은 필드를 직접 못 바꿉니다. 꺼내서 바꾸고 다시 넣거나 포인터 맵을 씁니다.
	type P struct{ N int }
	pm := map[string]*P{"a": {N: 1}}
	pm["a"].N++
	fmt.Println("pointer map:", pm["a"].N)
	fmt.Println()
}

// ── 문자열과 룬 ─────────────────────────────────────────────────────────────
// string 은 바이트의 불변 시퀀스입니다. len 은 바이트 수. 한글은 글자당 3바이트.
// 글자 단위로 다루려면 rune([]rune 또는 range).
func stringsAndRunes() {
	s := "가격 ₩1,000"
	fmt.Println("bytes:", len(s), " runes:", utf8.RuneCountInString(s))

	for i, r := range s { // range 는 룬 단위로 돌고 i 는 바이트 오프셋
		if i > 6 {
			break
		}
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	rs := []rune(s)
	fmt.Println("first two runes:", string(rs[:2]))

	// strings 패키지가 대부분을 해결합니다.
	fmt.Println(strings.Fields("  a  b \n c "), strings.Split("a,b,c", ","), strings.TrimSpace("  x  "))
	fmt.Println(strings.HasPrefix(s, "가격"), strings.Contains(s, "₩"), strings.ReplaceAll(s, ",", ""))

	var b strings.Builder // 문자열을 반복해서 이어 붙일 땐 Builder
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, "%d-", i)
	}
	fmt.Println(b.String())
	fmt.Println()
}

// ── 정렬 ───────────────────────────────────────────────────────────────────
func sorting() {
	xs := []int{3, 1, 2}
	slices.Sort(xs)
	fmt.Println("sorted:", xs)

	type item struct {
		name  string
		price int
	}
	items := []item{{"b", 20}, {"a", 10}, {"c", 10}}
	slices.SortFunc(items, func(x, y item) int { // 비교 함수: 음수/0/양수
		if x.price != y.price {
			return x.price - y.price
		}
		return strings.Compare(x.name, y.name)
	})
	fmt.Println("by price then name:", items)

	// 옛 방식(sort.Slice)도 여전히 많이 보입니다.
	sort.Slice(items, func(i, j int) bool { return items[i].name > items[j].name })
	fmt.Println("by name desc:", items)
	fmt.Println()
}

// ── 제네릭 ─────────────────────────────────────────────────────────────────
// 타입 파라미터로 "어떤 타입이든" 되는 함수/타입을 만듭니다. 제약으로 범위를 좁힙니다.
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

type Number interface{ ~int | ~int64 | ~float64 }

func Max[T Number](xs ...T) T {
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// 제네릭 타입: 어떤 T 의 스택
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func generics() {
	prices := []int{1890000, 1690000, 1590000}
	fmt.Println(Map(prices, func(p int) string { return fmt.Sprintf("₩%d", p) }))
	fmt.Println(Filter(prices, func(p int) bool { return p < 1700000 }))
	fmt.Println("max:", Max(prices...), Max(1.5, 2.5))

	var st Stack[string]
	st.Push("a")
	st.Push("b")
	v, _ := st.Pop()
	fmt.Println("popped:", v)
}
