# go-study

Go 연습장. 제품 저장소([pricewatch](https://github.com/CodingToastBread/pricewatch))와 분리해서,
실험·예제·막혔던 것들을 여기에 쌓습니다.

## 규칙

- 폴더 하나 = 주제 하나. 각 폴더는 독립된 모듈(`go.mod`)입니다.
- 각 폴더의 `README.md` 에 세 줄: **해보려던 것 / 막힌 것 / 알게 된 것**.
- 아래 목차에 한 줄씩 추가합니다. 이것보다 복잡하게 만들지 않습니다.

## 목차

| 폴더 | 무엇 | 상태 |
|---|---|---|
| `hello-world/` | 첫 실행 | 완료 |
| `basics/` | Go 기본기 한 장. `go run .` 으로 12개 블록이 순서대로 찍힘 | 읽는 중 |
| `fetchall/` | 몸풀기: URL 여러 개를 동시에 가져오는 CLI (goroutine·채널·context·httptest) | 진행 중 |

## 실행

```bash
cd <폴더> && go run . && go test -race ./...
```
