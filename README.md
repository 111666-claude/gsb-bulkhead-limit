# bulkhead-limit

按租户做并发舱壁：超出并发的请求排队，队列满则拒绝，释放时按先来后到放行。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/bulkhead --sample=idem
go run ./cmd/bulkhead --sample=queuecap
go run ./cmd/bulkhead --sample=release
go run ./cmd/bulkhead --sample=fair
go run ./cmd/bulkhead --sample=work
```

## 口径

- **准入**：租户在用并发数小于上限时直接准入，并占用一个槽位。
- **排队**：达到上限的请求进入该租户的等待队列；队列长度达到上限时返回 `rejected`。
- **幂等**：同一请求 ID 重复申请返回第一次的结果，不重复占槽位、不重复入队。
- **释放**：释放后腾出槽位，并按先来后到放行最早等待的、且其租户已有空位的请求。

## 不变量

- 任一租户的在用并发不超过上限，等待队列长度不超过队列上限。
- 请求 ID 与结果一一对应。
- `scanned` 不随请求数平方级放大：两千次申请的 `scanned` 不超过 6000。

## 输出契约

场景打印一行 JSON，键随场景而定。`--sample=work` 打印 `{"requests": N, "scanned": N}`。
