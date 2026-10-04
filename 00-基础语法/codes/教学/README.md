# 教学示例代码

配合 [`../../教学文稿/`](../../教学文稿) 与 [`../../教案/`](../../教案) 使用的可运行示例，由原 `codes/demo_1.go` ~ `demo_26.go` 重新整理而成。

## 为何改成子目录结构

原 `codes/` 目录下 26 个文件**同属一个 package 且各有 `func main()`**，因此：

```bash
go build ./...   # ❌ 报错：main redeclared in this block
```

Go 要求同一目录只能有一个 `package main` 的入口。本目录改为**一篇一个子目录**，既保留 `go run` 的直观体验，又能做全量编译校验（校验方式见文末）。

## 环境

- Go 1.22+（教案基线为 Go 1.27，推荐 1.24+）
- 本目录有独立 `go.mod`（module `demo`），不依赖仓库其他目录，**也不使用任何第三方库**，可离线运行

## 常用命令

```bash
cd 00-基础语法/codes/教学

go vet ./...                              # 静态检查全部示例
gofmt -l .                                # 检查格式（应无输出）
go build ./...                            # 全量编译
go test ./...                             # 全部测试
go run ./s02_var                          # 运行单个示例
go test -bench . -benchmem ./s16_bench    # 跑基准测试
go test -bench . -benchmem ./s17_pool     # sync.Pool 复用 vs 每次 new
go build -gcflags '-m -l' ./s17_escape    # 逃逸分析
go run -race ./s15_syncmap                # 竞态检测
```

> ⚠️ `s16_bench/` 是**测试包**（`package bench`，只有 `crypto.go` + `crypto_test.go`），
> 没有 `func main()`，**只能用 `go test` 运行**，`go run ./s16_bench` 会失败。

## 目录索引

| 目录 | 对应文稿 / 教案 | 演示内容 | 运行方式 |
|---|---|---|---|
| `s01_hello/` | 01 | Hello World | `go run` |
| `s02_var/` | 02 | 常量/变量声明、零值、iota、四种 fmt 输出 | `go run` |
| `s03_array/` | 03 一 | 一维/二维数组、三大注意事项 | `go run` |
| `s04_slice/` | 03 二 | 声明、截取、三索引切片、append 扩容、三种删除 | `go run` |
| `s05_struct/` | 04 一 | 命名/匿名结构体、嵌套、接收者、JSON tag | `go run` |
| `s06_map/` | 03 三 | 六种声明、nil map 陷阱、编辑删除、有序遍历 | `go run` |
| `s07_loop/` | 05 | 三种容器遍历、break/continue/goto/switch | `go run` |
| `s08_func/` | 06 一~三 | 传值传指针、命名返回值、错误处理、不定参数、闭包 | `go run` |
| `s08_sign/` | 06 四 | MD5 + 生成签名（`strings.Builder` 拼接） | `go run` |
| `s08_defer/` | 06 四 | defer 经典题 + 五个坑 + 资源释放 | `go run` |
| `s09_interface/` | 07 一 | 隐式实现、编译期断言、接口组合、类型断言 | `go run` |
| `s10_option/` | 07 二 | Options 模式、结构体参数对比、**sync.Pool 反例**（见下） | `go run` |
| `s11_timeutil/` | 10 一二 | RFC3339 转 CST、time 常用操作、定时器 | `go run` |
| `s12_json/` | 04 三四 | 强类型/弱类型/匿名内嵌摊平、科学计数法与 2^53 精度边界 | `go run` |
| `s13_chan/` | 08 一 | channel 阻塞、close 规则、方向化、select、Worker Pool | `go run` |
| `s14_waitgroup/` | 08 二 | 正确用法、三个坑、context 取消、并发收集结果 | `go run` |
| `s15_syncmap/` | 08 三 | Store/Load/Range、六个方法、RWMutex 对比、-race | `go run` |
| `s16_bench/` | 09 一二 | MD5/AES/JWT 基准、字符串三写法、map vs struct | **`go test`** |
| `s17_escape/` | 09 三 | interface 字段/返回指针/大对象三种逃逸 | `go run` |
| `s17_pool/` | 09 四 | sync.Pool 复用临时对象、reset 必要性、基准对比 | `go run` + `go test` |

## 本次审阅修正的代码缺陷

逐条依据见 [`../../教案/00-审阅报告.md`](../../教案/00-审阅报告.md)（条目 P1~P9）：

| 目录/文件 | 修正 |
|---|---|
| `s16_bench/` | `main.go`→`crypto.go`、`main_test.go`→`crypto_test.go`（文件名与 `package bench` 及文件头注释对齐）；删除未使用的 `bytes`/`fmt` import 与两行 `var _ =` 占位 |
| `s13_chan/main.go` | 删除 `make(<-chan string)` / `make(chan<- string)` 的坏示范，改为"双向创建 + 转换出只读/只写视图 + 方向约束写在函数签名上" |
| `s12_json/main.go` | 删除无效的 `json:",squash"` 标签（`encoding/json` 不认识它）；修正"雪花 ID 精度已丢失"的错误示例，改为 **2^53 边界对照**（`1759482245000000` 只是显示问题，`7300000000000000001` 才真丢精度） |
| `s17_pool/` | 修正"每轮 GC 都会清空 pool"→ victim cache 两个 GC 周期；删除 `time.Now()` 手写微基准，新增 `pool_test.go`（`testing.B` + `allocs/op`） |
| `s08_sign/main.go` | `str + fmt.Sprintf(...)` 改为 `strings.Builder` + `Grow`；签名前缀统一为 `k=v&k=v` |
| `s05_struct/main.go` | 修正"忽略 Age"的矛盾注释；零值演示移到赋值之前 |
| `s17_escape.exe` | 已删除（2.3 MB 编译产物不应入库）；根目录 `.gitignore` 已补齐 |

## 待讨论的反例（教学用，勿模仿）

- **`s10_option/friend/option.go` 用 `sync.Pool` 池化一个 5 字段的 `option`**：对象约 48 字节，分配成本近乎为零；而每次 `WithXxx(...)` 都会为返回的闭包分配一次，池化省不掉。代价是必须维护 `reset()`，漏一个字段就是脏数据串号。
  课堂请用 `s17_pool` 的基准方法验证它是否真的有效：`go test -bench . -benchmem ./s17_pool`。
  详见教案 07 篇 §5.5（审阅条目 A16）。

## 基准数据（历史数据，仅供参考）

`go test -bench . -benchmem -benchtime 100x ./s16_bench`（Intel i7-6700HQ @ 2.60GHz）：

```text
BenchmarkMD5-8             297.0 ns/op       32 B/op       1 allocs/op
BenchmarkAES-8   (优化前)  3909 ns/op     2272 B/op      12 allocs/op   // 每次 Encrypt 都 NewCipher
BenchmarkAES-8   (优化后)  1029 ns/op     1248 B/op      10 allocs/op   // 缓存 cipher.Block（快 3.8 倍）
BenchmarkJWT-8             6875 ns/op     2032 B/op      27 allocs/op
BenchmarkStringOp1-8     1200190 ns/op   3212604 B/op    999 allocs/op   // +
BenchmarkStringOp2-8     1434781 ns/op   3233268 B/op   2002 allocs/op   // fmt.Sprintf
BenchmarkStringOp3-8        4509 ns/op      6144 B/op       1 allocs/op   // strings.Builder
BenchmarkMapOperation-8    81.00 ns/op        0 B/op       0 allocs/op
BenchmarkStructOperation-8 2.000 ns/op        0 B/op       0 allocs/op
```

> AES 优化前后的对比是**同一份代码的两个版本**实测：缓存 `cipher.Block`（密钥扩展只做一次）后，单次加密从 3909ns 降到 1029ns。这就是 09 篇「复用重量级对象」的实证；CBC 模式对象因有内部状态仍须每次新建（见 `crypto.go` 注释）。

> ⚠️ **这些数字只是某一台机器上的历史采样，不要当作结论**：
> - `-benchtime 100x` 只跑 100 次，`MapOperation`/`StructOperation` 两行的噪声大于信号；
> - 跨机器、跨 Go 版本不可直接比较；
> - **结论只看数量级与 `allocs/op`**。请在自己的环境用 `-benchtime=1s -count=5` 重测，并用 `benchstat` 判断差异是否显著。

## 说明

- `s15_syncmap/main.go` 中的 `raceMap()` 默认不调用，取消注释可复现 `fatal error: concurrent map read and map write`
- `s13_chan/main.go` 中的无缓冲 channel 死锁示例已注释（会终止程序），详见函数内注释
- `s12_json/main.go` 的弱类型解析用标准库手动实现（原教程用 `github.com/mitchellh/mapstructure`），文稿中保留了 mapstructure 版本供对照

## 校验状态

最近一次全量校验记录（审阅条目 B12 要求实证）：

| 校验时间 | Go 版本 | 命令 | 结果 |
|---|---|---|---|
| 2026-10-04 | go1.25.1 windows/amd64 | `gofmt -l .` | ✅ 无输出（全部已格式化） |
| 2026-10-04 | go1.25.1 windows/amd64 | `go vet ./...` | ✅ 无告警 |
| 2026-10-04 | go1.25.1 windows/amd64 | `go build ./...` | ✅ 通过 |
| 2026-10-04 | go1.25.1 windows/amd64 | `go test ./...` | ✅ ok（s16_bench、s17_pool） |
| 2026-10-04 | go1.25.1 windows/amd64 | `go run ./s01_hello` ~ `./s17_pool`（逐个冒烟） | ✅ 输出符合文稿预期（含 defer 经典题 B D C A） |
| 2026-10-04 | go1.25.1 windows/amd64 | `go test -race ./s15_syncmap` | ⚠️ 环境不支持（`-race` 需 64 位 gcc/CGO，本机缺 gcc），**代码无已知竞态**，建议在 CI/Linux 上复跑 |

复跑命令：

```bash
cd 00-基础语法/codes/教学
gofmt -l . && go vet ./... && go build ./... && go test ./... && go test -race ./s15_syncmap
```

建议把这条命令加进 CI（`.github/workflows/ci.yml`），并在本文件追加新的校验记录。
