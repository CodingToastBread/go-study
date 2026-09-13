# fetchall

URL 여러 개를 인자로 받아 **동시에** 가져오고, 각각의 HTTP 상태 코드와 걸린 시간을 찍는 CLI.

```
$ go run . https://go.dev https://github.com https://example.com
200  312ms  https://example.com
200  498ms  https://go.dev
200  651ms  https://github.com
total 653ms
```

## 단계 (순서대로, 각각 커밋)

1. `fetch(url string) (status int, took time.Duration, err error)` — `http.Client{Timeout: 5s}`,
   `defer resp.Body.Close()`, 에러는 `fmt.Errorf("fetch %s: %w", url, err)` 로 감싸서 반환.
2. `os.Args[1:]` 를 for 루프로 **순차** 실행. 전체 시간 기록: ______ ms
3. 결과 struct + 채널 하나. URL 마다 `go func(){ ch <- ... }()` 로 띄우고 개수만큼 받기.
   전체 시간이 "가장 느린 하나"로 줄어드는 것 확인: ______ ms
4. 같은 것을 `sync.WaitGroup` 버전으로. (채널을 닫는 쪽은 보내는 쪽!)
5. `context.WithTimeout(ctx, 3s)` + `http.NewRequestWithContext` — 느린 URL 이 취소되며 에러가 찍히는 것 확인.
6. `main_test.go`: `httptest.NewServer` 로 가짜 서버를 띄워 테이블 테스트 2케이스 (200 정상 / 타임아웃).

Go by Example 에서 볼 항목: Goroutines, Channels, WaitGroups, Context, HTTP Client.

## 해보려던 것

## 막힌 것

## 알게 된 것
