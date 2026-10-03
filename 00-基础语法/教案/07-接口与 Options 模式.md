# 07 ｜ 接口与 Options 模式

> **课时**：2 学时（90 分钟）｜**前置**：06
> **对应讲义**：`../教学文稿/07-接口与 Options 模式.md`｜**对应示例**：`s09_interface/` `s10_option/`
> **本篇勘误**：A15（废弃 API）、A16（sync.Pool 反模式）、A17（不可达 case）

---

## 1. 教学目标

### 1.1 知识目标
1. 能说出 Go 接口是**隐式实现**（无 `implements` 关键字），并解释它带来的解耦价值。
2. 能解释 `var _ I = (*T)(nil)` 的作用与生效时机（编译期）。
3. 能说出 Options 模式的**收益**（向后兼容）与**代价**（失去编译期参数检查）。

### 1.2 能力目标
1. 用"私有 struct + 导出接口 + 构造函数"封装一个模块。
2. 用 `opts ...Option` 实现可扩展的配置 API。
3. 会用类型断言与类型 switch，并知道**断言失败会 panic**。

### 1.3 素养目标
1. **接口要小**：`io.Reader` 只有一个方法。接口越大，实现成本越高，复用性越差。
2. **不过度设计**：配置项少于 3 个时，结构体参数比 Options 模式更好。

## 2. 重点与难点

| | 内容 | 处理方式 |
|---|---|---|
| **重点** | 隐式实现 + 编译期断言 | 现场删掉一个方法，看编译报错 |
| **重点** | Options 模式的写法与选型 | 与"结构体参数"并排对比 |
| **难点** | 方法集规则（值接收者 vs 指针接收者对接口的影响） | 现场制造"没实现接口"的困惑 |
| **难点** | 类型 switch 的顺序匹配（A17） | 现场演示不可达 case |
| **难点** | 什么时候**不**该用 `sync.Pool`（A16） | 用分配数量级说话 |

## 3. 课前准备
投影；准备 `grpc.NewClient` 的官方文档截图（显示 `Dial` 已 deprecated）。

## 4. 教学流程（90 分钟）

| 时间 | 环节 | 内容 |
|---|---|---|
| 00–08 | 导入 | 「为什么 Go 里没有 `implements`？」——从 Java 的接口痛点切入 |
| 08–30 | 讲授 | 隐式实现、编译期断言、接口组合、空接口 |
| 30–42 | 讲授 | 方法集与接收者（接口实现的隐蔽陷阱） |
| 42–52 | 讲授 | 类型断言与类型 switch（A17） |
| 52–58 | 休息 | |
| 58–80 | 讲授+演示 | 不定参数 → Options 模式 → 何时别用 |
| 80–90 | 练习+小结 | §7 + §5.5 |

## 5. 讲授要点（含板书）

### 5.1 板书一：隐式实现与封装

```go
// study/study.go
package study

import "errors"

// 编译期断言：*study 未实现 Study 就编译不过
var _ Study = (*study)(nil)

type Study interface {
	Listen(msg string) string
	Speak(msg string) string
}

type study struct{ name string }        // 私有：外部拿不到具体类型

func (s *study) Listen(msg string) string { return s.name + " 听 " + msg }
func (s *study) Speak(msg string) string  { return s.name + " 说 " + msg }

// 构造函数返回接口，并承担参数校验
func New(name string) (Study, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	return &study{name: name}, nil
}
```

**三件套的价值**
1. `var _ Study = (*study)(nil)`：接口一改，**编译期**立刻报错，而不是等到运行时。
2. `study` 私有：把字段名从 `name` 改成 `userName`，外部调用方**零改动**。
3. `New` 返回接口：调用方只看到方法，看不到实现。

> **接口设计原则：接口越小越好。** `io.Reader` 只有一个 `Read`，所以全宇宙都能实现它。四个方法的接口已经偏大；十个方法的接口基本无法复用。

**接口组合**
```go
type Reader interface{ Read(p []byte) (int, error) }
type Writer interface{ Write(p []byte) (int, error) }
type ReadWriter interface {
	Reader
	Writer
}
var _ ReadWriter = (*os.File)(nil)
```

### 5.2 板书二：方法集陷阱（高频面试点）

| 类型 | 方法集 |
|---|---|
| `T` | 所有**值接收者**方法 |
| `*T` | 值接收者 + 指针接收者方法 |

```go
type Counter struct{ n int }
func (c Counter) Get() int  { return c.n }
func (c *Counter) Add()     { c.n++ }

var _ interface{ Add() } = &Counter{}   // ✅
var _ interface{ Add() } = Counter{}    // ❌ Counter 的方法集里没有 Add
```
> 这就是 04 篇强调"同一类型接收者风格统一"的原因：混用会让**值**在需要接口的地方静默失败。
> 另注：可寻址的值调用指针方法时编译器会自动取地址（`c.Add()` 等价 `(&c).Add()`），但**存进接口时不会**。

### 5.3 板书三：类型断言与类型 switch

```go
var i any = "go"

s := i.(string)          // 单返回值：断言失败 panic
s, ok := i.(string)      // 双返回值：安全，永远优先用这个

switch v := i.(type) {
case int:
	_ = v + 1            // v 是 int
case string:
	_ = len(v)           // v 是 string
default:
	fmt.Printf("未知 %T\n", v)
}
```

**⚠️ 类型 switch 按书写顺序匹配，先窄后宽（勘误 A17）**
```go
switch v := x.(type) {
case fmt.Stringer:        // 若 *User 实现了 String()，这里就吃掉了它
	fmt.Println(v.String())
case *User:               // ❌ 永不可达
	fmt.Println(v.Name)
}
```
> 讲义里的示例把"宽接口 case"写在前面，导致 `case *study` 是死代码；而且 `study` 是包内私有类型，示例放到 `main` 包根本编译不过。**先窄后宽**，并且注意示例的作用域。

**空接口的代价**
- `any`（`interface{}`）会丢失编译期类型检查；
- 把非指针值装进接口通常**触发逃逸**（09 篇）；
- 这也是 `json.Unmarshal` 到 `map[string]any` 时数字全变 `float64` 的根因（04 篇）。
> 字段类型确定时**务必用具体类型**。Go 1.18+ 优先考虑**泛型**而不是 `any`。

### 5.4 板书四：从 `grpc.NewClient` 学 Options 模式

```go
// 现代写法（勘误 A15：Dial 与 WithInsecure 均已废弃）
import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

conn, err := grpc.NewClient("127.0.0.1:8000",
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

**为什么长这样**：`opts ...DialOption` 是**不定参数 + 函数类型**的组合。
```go
type Option func(*option)

func WithTimeout(d time.Duration) Option {
	return func(o *option) { o.timeout = d }   // 闭包捕获参数
}

func New(addr string, opts ...Option) (*Client, error) {
	o := defaultOption()          // 1. 先设默认值
	for _, f := range opts {
		f(o)                      // 2. 逐个覆盖
	}
	if err := o.validate(); err != nil {   // 3. 集中校验
		return nil, err
	}
	return &Client{opt: o}, nil
}
```

**完整可运行示例**
```go
// friend/option.go
package friend

import "time"

type Option func(*option)

type option struct {
	sex     int
	age     int
	timeout time.Duration
}

func defaultOption() *option {
	return &option{sex: 0, age: 0, timeout: 3 * time.Second}
}

func WithSex(sex int) Option        { return func(o *option) { o.sex = sex } }
func WithAge(age int) Option        { return func(o *option) { o.age = age } }
func WithTimeout(d time.Duration) Option {
	return func(o *option) { o.timeout = d }
}
```
```go
// friend/friend.go
package friend

import (
	"errors"
	"fmt"
	"strings"
)

func Find(where string, opts ...Option) (string, error) {
	o := defaultOption()
	for _, f := range opts {
		f(o)
	}
	if o.age < 0 || o.age > 150 {
		return "", errors.New("age 超出合法范围")
	}

	var sb strings.Builder          // 勘误 B11：不要用 + 累加
	sb.WriteString("从 " + where + " 找朋友\n")
	if o.sex == 1 {
		sb.WriteString("性别：女性\n")
	}
	if o.age != 0 {
		fmt.Fprintf(&sb, "年龄：%d岁\n", o.age)
	}
	fmt.Fprintf(&sb, "超时：%s\n", o.timeout)
	return sb.String(), nil
}
```

**Options 模式选型**

| | Options 模式 | 结构体参数 |
|---|---|---|
| 新增配置项 | ✅ 不破坏已有调用 | ❌ 调用方要改（除非用零值默认） |
| 参数检查 | ❌ 推迟到运行时 | ✅ 编译器检查字段名 |
| 可读性 | 配置项多时更长 | 一目了然 |
| 适用 | **已发布、调用方众多**的库 | 内部代码、配置项 < 3 个 |

```go
// 配置项少时，这个更好：
type FindOptions struct {
	Sex int
	Age int
}
func Find(where string, opts FindOptions) (string, error)
```

### 5.5 板书五：什么时候**不**该用 sync.Pool（勘误 A16）

讲义的 Options 示例用 `sync.Pool` 复用一个 5 字段的 `option` 结构体，这是**反模式**：

| 维度 | 事实 |
|---|---|
| 对象大小 | 5 个字段 ≈ 48 字节，分配成本近乎为零 |
| 真正的分配来源 | 每次 `WithXxx(...)` 都会为**返回的闭包**分配一次——池化 option 省不掉 |
| 引入的风险 | 必须维护 `reset()`；漏一个字段就是**上一个请求的数据串到下一个请求** |
| 净收益 | ≈ 0，可读性下降 |

**sync.Pool 的正确场景**（09 篇详述）：`bytes.Buffer`、大 `[]byte`、编解码器等**分配代价高、生命周期短、高频**的对象。
**判据**：先用 `-benchmem` 量出 `allocs/op`，确认"这个对象的分配确实是热点"，再上池。

### 5.6 小结（口述六句）
1. 接口隐式实现，`var _ I = (*T)(nil)` 做编译期断言。
2. 私有 struct + 构造函数返回接口 = 封装。
3. `T` 的方法集不含指针方法——接收者风格要统一。
4. 类型断言优先用双返回值；类型 switch **先窄后宽**。
5. `opts ...Option` 就是 Options 模式；它的价值是**向后兼容**，代价是失去编译期检查。
6. 不要滥用 `any`，也不要滥用 `sync.Pool`。

## 6. 演示脚本

```go
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ---------- 接口 ----------

type Animal interface{ Sound() string }

type Dog struct{ Name string }
func (d *Dog) Sound() string { return d.Name + ": 汪" }

type Cat struct{ Name string }
func (c *Cat) Sound() string { return c.Name + ": 喵" }

var _ Animal = (*Dog)(nil)   // 编译期断言
var _ Animal = (*Cat)(nil)

// ---------- Options 模式 ----------

type Option func(*option)
type option struct {
	sex     int
	age     int
	timeout time.Duration
}

func defaultOption() *option { return &option{timeout: 3 * time.Second} }
func WithSex(s int) Option   { return func(o *option) { o.sex = s } }
func WithAge(a int) Option   { return func(o *option) { o.age = a } }
func WithTimeout(d time.Duration) Option {
	return func(o *option) { o.timeout = d }
}

func Find(where string, opts ...Option) (string, error) {
	o := defaultOption()
	for _, f := range opts {
		f(o)
	}
	if o.age < 0 || o.age > 150 {
		return "", fmt.Errorf("age 非法: %d", o.age)
	}
	var sb strings.Builder
	sb.WriteString("从 " + where + " 找朋友\n")
	if o.sex == 1 {
		sb.WriteString("性别：女性\n")
	}
	if o.age != 0 {
		fmt.Fprintf(&sb, "年龄：%d岁\n", o.age)
	}
	fmt.Fprintf(&sb, "超时：%s\n", o.timeout)
	return sb.String(), nil
}

// ---------- 类型 switch 顺序 ----------

type Named interface{ Name() string }
type User struct{ n string }
func (u *User) Name() string { return u.n }

func main() {
	// 1) 接口多态
	for _, a := range []Animal{&Dog{"旺财"}, &Cat{"咪咪"}} {
		fmt.Println("1)", a.Sound())
	}

	// 2) 编译期断言让错误提前暴露：
	//    把 Dog 的 Sound() 改名，这一行立刻编译不过

	// 3) Options
	s, err := Find("附近的人", WithSex(1), WithAge(30), WithTimeout(5*time.Second))
	fmt.Println("3)", s, err)

	_, err = Find("附近的人", WithAge(200))
	fmt.Println("   校验失败:", err)

	// 4) 先宽后窄 → 窄类型不可达
	var x Named = &User{"Tom"}
	switch v := x.(type) {
	case Named:
		fmt.Println("4) 命中 Named:", v.Name())
	case *User:
		fmt.Println("   永远不会到这里")
	}

	// 5) 安全断言
	var i any = "go"
	if s, ok := i.(string); ok {
		fmt.Println("5) 断言成功:", s)
	}
	if _, ok := i.(int); !ok {
		fmt.Println("   不是 int（不 panic）")
	}
	_ = errors.New
}
```

> 对应仓库示例：`s09_interface/`、`s10_option/`。注意 `s10_option` 里的 `sync.Pool` 是**待讨论的反例**（§5.5）。

## 7. 课堂练习（含参考答案）

**练习 1** 下面代码为什么编译不过？两种改法。
```go
type Storer interface{ Save() error }
type FileStore struct{}
func (f FileStore) Save() error { return nil }
var s Storer = &FileStore{}   // ？
```
> **答**：这段其实**能**编译——`*FileStore` 的方法集包含值接收者方法 `Save`。真正编译不过的是反过来：`func (f *FileStore) Save()` 时写 `var s Storer = FileStore{}`。两种改法：① 统一用指针接收者并传 `&FileStore{}`；② 统一用值接收者。**关键是统一**，并加 `var _ Storer = (*FileStore)(nil)` 让编译器帮你盯。

**练习 2** 为 `Find` 增加 `WithHobby(h string)`，要求 hobby 为空字符串时不输出该行。说明为什么"新增选项不会破坏已有调用方"。
> **答**：新增 `WithHobby` 后，老代码不传它就走 `defaultOption()` 的零值，行为不变；如果用结构体参数，新增字段虽然也不强制，但**位置参数**式的 API（`Find(where, 1, 30, "")`）就必须改所有调用点。Options 模式的兼容性来自"每个选项自带默认值"。

**练习 3** 把下面函数改造成 Options 模式，并写出默认值与校验：
```go
func Dial(addr string, timeout time.Duration, retries int, insecure bool) error
```
> **答**：
> ```go
> type DialOption func(*dialOption)
> type dialOption struct {
> 	timeout  time.Duration
> 	retries  int
> 	insecure bool
> }
> func defaultDialOption() *dialOption {
> 	return &dialOption{timeout: 5 * time.Second, retries: 3, insecure: false}
> }
> func WithDialTimeout(d time.Duration) DialOption { return func(o *dialOption) { o.timeout = d } }
> func WithRetries(n int) DialOption              { return func(o *dialOption) { o.retries = n } }
> func WithInsecure() DialOption                  { return func(o *dialOption) { o.insecure = true } }
> func Dial(addr string, opts ...DialOption) error {
> 	o := defaultDialOption()
> 	for _, f := range opts { f(o) }
> 	if o.timeout <= 0 { return errors.New("timeout 必须为正") }
> 	if o.retries < 0 { return errors.New("retries 不能为负") }
> 	// ...
> 	return nil
> }
> ```

**练习 4** 讨论题：`s10_option` 用 `sync.Pool` 复用 `option`。请设计一个实验，用数据证明它是"无效优化"。
> **答**：写两个基准：`BenchmarkFindNoPool`（每次 `defaultOption()` 新建）与 `BenchmarkFindPool`（Get/Put + reset）。用 `go test -bench . -benchmem` 对比 `ns/op` 与 `allocs/op`。预期：两者差距在噪声范围内，且 `allocs/op` 都被 `WithXxx` 的闭包分配主导。若差距不显著，就应删掉池（复杂度和脏数据风险不值得）。

## 8. 勘误与易错预警

| 编号 | 讲义原状 | 教案处理 |
|---|---|---|
| **A15** | 示例用 `grpc.Dial` + `grpc.WithInsecure()` | 改为 `grpc.NewClient` + `WithTransportCredentials(insecure.NewCredentials())`；补一句"`NewClient` 默认解析器是 `dns`，`Dial` 是 `passthrough`，迁移直连场景要注意" |
| **A16** | 用 `sync.Pool` 池化 5 字段 option | 改为**反例讨论**（§5.5），并布置练习 4 用基准数据说话 |
| **A17** | 类型 switch 里 `case *study` 不可达且跨包不可见 | 换成 `Named` / `*User` 的可编译示例，并强调**先窄后宽** |
| — | 学员写 `interface{}` | Go 1.18+ 用 `any`（等价但更短）；能用具体类型/泛型就不用 `any` |
| — | 学员给接口加方法"顺便扩展" | 接口一旦发布就难改；扩展用**新接口 + 组合**（如 `io.ReaderFrom`） |
| — | 学员把接口定义在实现方包里 | Go 惯例：**接口定义在使用方**（consumer side），实现方只管提供方法 |

## 9. 课后作业（分层）

- **基础**：定义 `Notifier` 接口（`Send(msg string) error`），实现 `EmailNotifier` 与 `LogNotifier`，写一个 `Broadcast(ns []Notifier, msg string) error`（用 `errors.Join` 汇总错误）。加 `var _ Notifier = (*EmailNotifier)(nil)`。
- **进阶**：把 06 篇的 `CreateSign` 改造成可配置的：`NewSigner(opts ...SignerOption)`，支持 `WithSecret`、`WithTimestamp`、`WithNonce`、`WithHash(算法)`。要求默认值可用（零配置也能签名），并写单测覆盖"加了 timestamp 后签名每次不同"。
- **挑战**：实现一个 `Repository` 接口 + 内存实现 + 一个 `WithCache(ttl)` 装饰器（返回同样实现 `Repository` 的包装类型）。说明这为什么是"接口 + 组合"优于"继承"的例子。

## 10. 考核点（可观测）

- [ ] 能解释隐式实现的价值，并写出编译期断言。
- [ ] 能说出 `T` 与 `*T` 的方法集差异，并解释一个由此产生的接口实现失败。
- [ ] 能徒手写出 Options 模式的骨架（默认值 → 应用 → 校验）。
- [ ] 能说出 Options 模式的一个收益与一个代价，并举出"不该用它"的场景。
- [ ] 能解释为什么池化小对象是无效优化。
- [ ] 进阶作业 `go test ./...` 通过。

## 11. 延伸阅读
- Go 谚语：*The bigger the interface, the weaker the abstraction.*
- `grpc.NewClient` 文档（注意 `Dial` 的 deprecated 说明）
- Rob Pike：*Go Proverbs*
- 下一课预告：并发——goroutine 不是线程、channel 的阻塞语义、WaitGroup 的三个坑、以及"每个 goroutine 都要自己 recover"。
