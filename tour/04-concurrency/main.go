// 04-concurrency — goroutine, 채널, 버퍼, 방향, select, 타임아웃, 타이머/티커,
//
//	워커풀, WaitGroup, Mutex, atomic, context, 레이트 리밋.
//
// 실행: go run ./04-concurrency
// 레이스 검사까지: go run -race ./04-concurrency
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	goroutines()
	channels()
	bufferedAndDirections()
	selectBasics()
	timersAndTickers()
	workerPool()
	waitGroups()
	mutexes()
	atomics()
	contexts()
	rateLimiting()
}

// ── goroutine ──────────────────────────────────────────────────────────────
// go f() 로 동시 실행. 아주 가볍습니다(수천 개도 OK). main 이 끝나면 전부 죽습니다.
func goroutines() {
	say := func(s string) { fmt.Print(s, " ") }
	go say("background")
	say("foreground")
	time.Sleep(10 * time.Millisecond) // 데모용. 실제로는 WaitGroup/채널로 기다립니다.
	fmt.Println()
	fmt.Println()
}

// ── 채널 ───────────────────────────────────────────────────────────────────
// goroutine 사이의 타입 있는 파이프. 버퍼 없는 채널은 보내는 쪽과 받는 쪽이 만나야 진행됩니다.
func channels() {
	msgs := make(chan string)
	go func() { msgs <- "ping" }() // 받는 쪽이 있을 때까지 여기서 대기
	fmt.Println(<-msgs)

	// 닫기: 보내는 쪽이 닫고, 받는 쪽은 range 로 닫힐 때까지 받습니다.
	nums := make(chan int)
	go func() {
		defer close(nums)
		for i := 1; i <= 3; i++ {
			nums <- i
		}
	}()
	for n := range nums {
		fmt.Print(n, " ")
	}
	fmt.Println()

	v, ok := <-nums // 닫힌 채널: 제로값 + false
	fmt.Println("closed read:", v, ok)
	fmt.Println()
}

// ── 버퍼와 방향 ─────────────────────────────────────────────────────────────
// 버퍼 채널은 버퍼가 찰 때까지 받는 쪽 없이도 보낼 수 있습니다.
// 함수 시그니처에 방향(chan<- 보내기만, <-chan 받기만)을 적으면 실수를 컴파일러가 잡습니다.
func producer(out chan<- int) {
	for i := 0; i < 3; i++ {
		out <- i * i
	}
	close(out)
}

func consumer(in <-chan int) (sum int) {
	for v := range in {
		sum += v
	}
	return
}

func bufferedAndDirections() {
	buf := make(chan string, 2)
	buf <- "a"
	buf <- "b" // 받는 쪽 없이도 OK (버퍼 2)
	fmt.Println(<-buf, <-buf)

	ch := make(chan int, 3)
	go producer(ch)
	fmt.Println("sum of squares:", consumer(ch))
	fmt.Println()
}

// ── select ─────────────────────────────────────────────────────────────────
// 여러 채널 중 준비된 하나를 실행. 타임아웃과 non-blocking 도 select 로.
func selectBasics() {
	c1, c2 := make(chan string), make(chan string)
	go func() { time.Sleep(10 * time.Millisecond); c1 <- "one" }()
	go func() { time.Sleep(20 * time.Millisecond); c2 <- "two" }()

	for i := 0; i < 2; i++ {
		select {
		case m := <-c1:
			fmt.Println("got", m)
		case m := <-c2:
			fmt.Println("got", m)
		}
	}

	// 타임아웃
	slow := make(chan string)
	go func() { time.Sleep(time.Second); slow <- "late" }()
	select {
	case m := <-slow:
		fmt.Println(m)
	case <-time.After(30 * time.Millisecond):
		fmt.Println("timeout")
	}

	// non-blocking: default 가 있으면 기다리지 않습니다.
	select {
	case m := <-slow:
		fmt.Println(m)
	default:
		fmt.Println("nothing ready")
	}
	fmt.Println()
}

// ── 타이머와 티커 ───────────────────────────────────────────────────────────
func timersAndTickers() {
	t := time.NewTimer(20 * time.Millisecond)
	<-t.C
	fmt.Println("timer fired")

	t2 := time.NewTimer(time.Second)
	if t2.Stop() {
		fmt.Println("timer stopped before firing")
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	done := time.After(35 * time.Millisecond)
	ticks := 0
loop:
	for {
		select {
		case <-ticker.C:
			ticks++
		case <-done:
			break loop // 라벨 break: 바깥 for 를 빠져나옴
		}
	}
	ticker.Stop()
	fmt.Println("ticks:", ticks)
	fmt.Println()
}

// ── 워커풀 ─────────────────────────────────────────────────────────────────
// jobs 채널 하나에 워커 N개가 range 로 붙으면 그게 워커풀입니다. 동시 실행 수 = 워커 수.
func workerPool() {
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	worker := func(id int) {
		for j := range jobs {
			time.Sleep(5 * time.Millisecond) // 일하는 척
			results <- j * 2
		}
	}
	for w := 1; w <= 3; w++ {
		go worker(w)
	}
	for j := 1; j <= 6; j++ {
		jobs <- j
	}
	close(jobs)

	sum := 0
	for i := 0; i < 6; i++ {
		sum += <-results
	}
	fmt.Println("worker pool sum:", sum)
	fmt.Println()
}

// ── WaitGroup ──────────────────────────────────────────────────────────────
// "전부 끝날 때까지" 기다리기. Add → 각 goroutine 에서 Done → Wait.
func waitGroups() {
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i) * 5 * time.Millisecond)
			fmt.Print("w", i, " ")
		}()
	}
	wg.Wait()
	fmt.Println("all done")

	// Go 1.25+: wg.Go(func(){...}) 가 Add/Done 을 대신합니다.
	var wg2 sync.WaitGroup
	total := 0
	var mu sync.Mutex
	for i := 1; i <= 3; i++ {
		wg2.Go(func() { mu.Lock(); total += i; mu.Unlock() })
	}
	wg2.Wait()
	fmt.Println("wg.Go total:", total)
	fmt.Println()
}

// ── Mutex ──────────────────────────────────────────────────────────────────
// 여러 goroutine 이 같은 데이터를 건드리면 락. 구조체 안에 mu 를 두는 게 관례.
// go test -race / go run -race 가 락 없는 곳을 잡아줍니다.
type Counter struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *Counter) Inc(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k]++
}

func mutexes() {
	c := Counter{m: map[string]int{}}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc("hits")
		}()
	}
	wg.Wait()
	fmt.Println("hits:", c.m["hits"])

	// 읽기가 많으면 RWMutex: 읽기는 동시에, 쓰기는 배타적으로.
	var rw sync.RWMutex
	rw.RLock()
	_ = c.m["hits"]
	rw.RUnlock()
	fmt.Println()
}

// ── atomic ─────────────────────────────────────────────────────────────────
// 숫자 하나만 안전하게 세면 되면 Mutex 대신 atomic.
func atomics() {
	var ops atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ops.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Println("ops:", ops.Load())
	fmt.Println()
}

// ── context ────────────────────────────────────────────────────────────────
// 취소·타임아웃·요청 범위 값을 첫 번째 인자로 흘려보냅니다.
// 오래 걸리는 일은 ctx.Done() 을 듣다가 신호가 오면 멈춥니다. cancel 은 반드시 호출(defer).
func slowWork(ctx context.Context, name string, takes time.Duration) error {
	select {
	case <-time.After(takes):
		fmt.Println(name, "finished")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", name, ctx.Err())
	}
}

func contexts() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_ = slowWork(ctx, "fast", 10*time.Millisecond)
	fmt.Println(slowWork(ctx, "slow", 100*time.Millisecond))

	// 수동 취소: 부모를 취소하면 자식도 전부 취소됩니다.
	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithTimeout(parent, time.Hour)
	defer cancelChild()
	cancelParent()
	<-child.Done()
	fmt.Println("child cancelled via parent:", child.Err())

	// 요청 범위 값: 키는 비공개 타입으로 (충돌 방지). 설정 전달용으로 남용하지 않기.
	type key struct{}
	withID := context.WithValue(context.Background(), key{}, "req-123")
	fmt.Println("value:", withID.Value(key{}))
	fmt.Println()
}

// ── 레이트 리밋 ─────────────────────────────────────────────────────────────
// 티커 하나로 "초당 N개" 를 만들 수 있습니다. 토큰 버킷은 버퍼 채널로.
func rateLimiting() {
	limiter := time.NewTicker(10 * time.Millisecond)
	defer limiter.Stop()
	start := time.Now()
	for i := 1; i <= 3; i++ {
		<-limiter.C // 10ms 마다 하나씩 통과
		fmt.Print("req", i, " ")
	}
	fmt.Println("in", time.Since(start).Round(10*time.Millisecond))

	bucket := make(chan struct{}, 3) // 최대 3개 버스트
	for i := 0; i < 3; i++ {
		bucket <- struct{}{}
	}
	burst := 0
	for len(bucket) > 0 {
		<-bucket
		burst++
	}
	fmt.Println("burst allowed:", burst)
}
