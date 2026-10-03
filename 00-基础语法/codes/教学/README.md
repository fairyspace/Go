# 教学示例代码

配合 [`../../教学文稿/`](../../教学文稿) 使用的可运行示例，由原 `codes/demo_1.go` ~ `demo_26.go` 重新整理而成。

## 为何改成子目录结构

原 `codes/` 目录下 26 个文件**同属一个 package 且各有 `func main()`**，因此：

```bash
go build ./...   # ❌ 报错：main redeclared in this block
```

Go 要求同一目录只能有一个 `package main` 的入口。本目录改为**一篇一个子目录**，既保留 `go run` 的直观体验，又支持全量编译校验。

## 环境

- Go 1.22+
- 本目录有独立 `go.mod`（module `demo`），不依赖仓库其他目录，**也不使用任何第三方库**，可离线运行

## 常用命令

```bash
cd 00-基础语法/codes/教学

go vet ./...                              # 静态检查全部示例
gofmt -l .                                # 检查格式
go run ./s02_var                          # 运行单个示例
go test -bench . -benchmem ./s16_bench    # 跑基准测试
go build -gcflags '-m -l' ./s17_escape   # 逃逸分析
go run -race ./s15_syncmap                # 竞态检测
```

## 目录索引

| 目录 | 对应文稿 | 演示内容 |
|---|---|---|
| `s01_hello/` | 01 | Hello World |
| `s02_var/` | 02 | 常量/变量声明、零值、iota、四种 fmt 输出 |
| `s03_array/` | 03 一 | 一维/二维数组、三大注意事项 |
| `s04_slice/` | 03 二 | 声明、截取、三索引切片、append 扩容、三种删除 |
| `s05_struct/` | 04 一 | 命名/匿名结构体、嵌套、接收者、JSON tag |
| `s06_map/` | 03 三 | 六种声明、nil map 陷阱、编辑删除、有序遍历 |
| `s07_loop/` | 05 | 三种容器遍历、break/continue/goto/switch |
| `s08_func/` | 06 一~三 | 传值传指针、命名返回值、错误处理、不定参数、闭包 |
| `s08_sign/` | 06 四 | MD5 + 生成签名 |
| `s08_defer/` | 06 四 | defer 经典题 + 五个坑 + 资源释放 |
| `s09_interface/` | 07 一 | 隐式实现、编译期断言、接口组合、类型断言 |
| `s10_option/` | 07 二 | Options 模式、sync.Pool 复用、结构体参数对比 |
| `s11_timeutil/` | 10 一二 | RFC3339 转 CST、time 常用操作、定时器 |
| `s12_json/` | 04 三四 | 强类型/弱类型/squash、科学计数法三方案 |
| `s13_chan/` | 08 一 | 并发vs并行、channel 阻塞、select、Worker Pool |
| `s14_waitgroup/` | 08 二 | 正确用法、三个坑、context 取消、并发收集结果 |
| `s15_syncmap/` | 08 三 | Store/Load/Range、六个方法、RWMutex 对比、-race |
| `s16_bench/` | 09 一二 | MD5/AES/JWT 基准、字符串三写法、map vs struct |
| `s17_escape/` | 09 三 | interface 字段/返回指针/大对象三种逃逸 |
| `s17_pool/` | 09 四 | sync.Pool 复用临时对象、reset 必要性 |

## 已验证的基准数据

`go test -bench . -benchmem -benchtime 100x ./s16_bench`（i7-6700HQ @ 2.60GHz）：

```text
BenchmarkMD5-8             297.0 ns/op       32 B/op       1 allocs/op
BenchmarkAES-8             3909 ns/op     2272 B/op      12 allocs/op
BenchmarkJWT-8             6875 ns/op     2032 B/op      27 allocs/op
BenchmarkStringOp1-8     1200190 ns/op   3212604 B/op    999 allocs/op   // +
BenchmarkStringOp2-8     1434781 ns/op   3233268 B/op   2002 allocs/op   // fmt.Sprintf
BenchmarkStringOp3-8        4509 ns/op      6144 B/op       1 allocs/op   // strings.Builder
BenchmarkMapOperation-8    81.00 ns/op        0 B/op       0 allocs/op
BenchmarkStructOperation-8 2.000 ns/op        0 B/op       0 allocs/op
```

> `allocs/op`（每次分配次数）比 `ns/op` 更能反映 GC 压力，是判断优化效果的**首要指标**。

## 说明

- `s15_syncmap/main.go` 中的 `raceMap()` 默认不调用，取消注释可复现 `fatal error: concurrent map read and map write`
- `s13_chan/main.go` 中的无缓冲 channel 死锁示例已注释（会终止程序），详见函数内注释
- `s12_json/main.go` 的弱类型解析用标准库手动实现（原教程用 `github.com/mitchellh/mapstructure`），文稿中保留了 mapstructure 版本供对照
