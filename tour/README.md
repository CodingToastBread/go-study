# tour — Go 한 바퀴

Go by Example 의 주제 순서를 따라, 실행되는 파일 다섯 개로 문법 전체를 한 번 훑습니다.
각 파일은 위에서 아래로 읽으면서 출력과 대조하고, 값을 바꿔 다시 돌려보는 용도입니다.
설명 문단은 없고 코드 + 주석만 있습니다. 한 파일에 20~30분, 전체 두 시간.

```bash
go run ./01-syntax        # 값, 변수, 상수, for, if/switch, 함수, 클로저, 포인터
go run ./02-collections   # 배열, 슬라이스(와 함정), 맵, 문자열/룬, 정렬, 제네릭
go run ./03-types         # 구조체, 메서드, 임베딩, 인터페이스, 에러(Is/As/Join), defer/panic/recover
go run -race ./04-concurrency  # goroutine, 채널, select, 타이머, 워커풀, WaitGroup, Mutex, atomic, context
go run ./05-stdlib        # fmt, time, JSON, 정규식, 파일, 환경변수, HTTP 서버/클라이언트
go test -v ./05-stdlib    # 테이블 테스트, httptest, 벤치마크, t.Helper
```

`basics/` 와의 관계: `basics/` 는 pricewatch 에 나오는 것만 깊게, `tour/` 는 전체를 넓게.
둘 다 훑고 나면 Go by Example 은 "기억 안 날 때 그 항목만 찾는 사전"으로 씁니다.

## 해보려던 것

## 막힌 것

## 알게 된 것
