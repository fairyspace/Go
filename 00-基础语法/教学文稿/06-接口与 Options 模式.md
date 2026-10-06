# 06 接口与 Options 模式

> 学完本篇你将能够：用隐式实现定义接口、用编译期断言保证实现完整、用 Options 模式写出可扩展的链式配置 API、看懂 `grpc.Dial`
> 的源码写法、知道何时不该用 Options 模式。

## 概述

本篇合并了原「结构体实现接口」和「学习 grpc.Dial 的写法」两篇教程。它们都在讲同一件事： **Go 的代码组织与可扩展性设计**。

---

## 一、结构体实现接口

在 Go 语言中， **一个 struct 实现了某个接口里的所有方法，就叫做这个 struct 实现了该接口**。

下面写一个 Demo 实现一下：先写一个 `Study interface{}`，里面需要实现 4 个方法 Listen、Speak、Read、Write，然后再写一个
`study struct{}`，去全部实现里面的方法，然后分享一下代码心得。

### 代码示例

```go
// s09_interface/study/study.go
package study

import "errors"

// 编译期断言：要求 *study 必须实现 Study
var _ Study = (*study)(nil)

type Study interface {
	Listen(msg string) string
	Speak(msg string) string
	Read(msg string) string
	Write(msg string) string
}

// 私有结构体：不想在其他地方被使用
type study struct {
	Name string
}

func (s *study) Listen(msg string) string { return s.Name + " 听 " + msg }
func (s *study) Speak(msg string) string  { return s.Name + " 说 " + msg }
func (s *study) Read(msg string) string   { return s.Name + " 读 " + msg }
func (s *study) Write(msg string) string  { return s.Name + " 写 " + msg }

// New 构造函数：返回接口而非具体类型
func New(name string) (Study, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	return &study{
		Name: name,
	}, nil
}
```

```go
// s09_interface/main.go
package main

import (
	"demo/study"
	"fmt"
)

func main() {
	name := "Tom"
	s, err := study.New(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(s.Listen("english"))
	fmt.Println(s.Speak("english"))
	fmt.Println(s.Read("english"))
	fmt.Println(s.Write("english"))
}
```

> 原始教程这里用的是 `github.com/pkg/errors`，此处改用标准库 `errors`，减少外部依赖。生产项目可按需换回。

输出：

```text
Tom 听 english
Tom 说 english
Tom 读 english
Tom 写 english
```

### 代码解释

#### 一、`var _ Study = (*study)(nil)`

```go
var _ Study = (*study)(nil)
```

要求 `*study` 去实现 `Study`，若 `Study` 接口被更改或未全部实现时， **在编译时就会报错**。

这是 Go 中标准的接口实现断言写法，`_` 表示丢弃该值，只保留类型检查。

#### 二、`type study struct` 定义为私有

```go
type study struct {
Name string
}
```

之所以定义为私有的结构体，是因为不想在其他地方被使用。比如后面将 `Name` 改成 `UserName`，只需要在本包内修改即可，外部调用方零改动。

#### 三、`New()` 返回接口而非具体类型

```go
func New(name string) (Study, error) {
if name == "" {
return nil, errors.New("name required")
}
return &study{
Name: name,
}, nil
}
```

在其他地方调用 `New()` 使用 `Study` 包时， **仅对外暴露了 4 个方法**，别人只管调用就好了，内部实现别人无需关心。

同时构造函数承担了 **参数校验**（返回 `error`）与 **可能的资源初始化**两个职责——这是 Go 生态中处理「创建可能失败的对象」的标准方式。

### 补充：接口组合

Go 的接口可以像结构体一样 **内嵌组合**：

```go
type Reader interface {
Read(p []byte) (n int, err error)
}

type Writer interface {
Write(p []byte) (n int, err error)
}

// 组合成新的接口
type ReadWriter interface {
Reader
Writer
}

// 一个类型可以同时实现多个接口
type File struct{ /* ... */ }

func (f *File) Read(p []byte) (int, error)  { /* ... */ }
func (f *File) Write(p []byte) (int, error) { /* ... */ }

var _ ReadWriter = (*File)(nil)
```

> 组合是接口的「能力聚合」，比让接口继承接口更符合 Go 的扁平风格。

### 补充：类型断言

```go
var s Study = &study{Name: "Tom"}

// 单返回值：失败会 panic
sw := s.(interface{ Write(msg string) string })

// 双返回值：安全
sw, ok := s.(interface{ Write(msg string) string })
if ok {
sw.Write("english")
}

// 类型 switch：更实用
// ⚠️ 类型 switch 按**书写顺序**匹配，第一个命中的 case 生效。
// 这里 *study 也实现了 Listen 方法，所以永远走第一个 case——
// 若把更宽的接口放在前面，窄类型 case 会变成永远不可达的死代码（常见 bug）。
switch v := s.(type) {
case interface{ Listen(string) string }:
fmt.Println(v.Listen("music"))
case *study: // 不可达：*study 必然命中第一个 case
fmt.Println("具体类型:", v.Name)
}
```

### 补充：空接口与鸭子类型

```go
// 空接口可以保存任何类型
var v interface{}
v = 1
v = "string"
v = []int{1, 2}

// 断言取出
if s, ok := v.(string); ok {
fmt.Println(s)
}
```

> ⚠️ **空接口 `interface{}` 会导致逃逸和性能下降**（见第 08 篇），且丧失编译期类型检查。字段类型确定时 **务必用具体类型**。
>
> 这也是为什么 `json.Unmarshal` 到 `map[string]interface{}` 会让所有数字变成 `float64`（第 04 篇）——因为 `interface{}`
> 只能存这几种基础类型。

---

## 二、学习 `grpc.NewClient(target string, opts ...DialOption)` 的写法

咱们平时是这样使用 gRPC 客户端的：

```go
// 现代 API（grpc-go 1.63+）
conn, err := grpc.NewClient("127.0.0.1:8000",
grpc.WithTransportCredentials(insecure.NewCredentials()),
grpc.WithChainStreamInterceptor(),
)

// ⚠️ 老项目里常见的两个 API 均已废弃：
//   grpc.Dial            → grpc.NewClient
//   grpc.WithInsecure()  → grpc.WithTransportCredentials(insecure.NewCredentials())
// 迁移注意：NewClient 默认名字解析器是 dns，而 Dial 是 passthrough，
// 直连 IP:PORT 的老代码迁移时需显式加 passthrough 或确认 DNS 可用。
```

咱们怎么能写出类似这样的调用方式，它是怎么实现的？

这篇文章咱们写一个 Demo，其实很简单，一步步往下看。

### 一、不定参数传递

`opts ...DialOption`，这个是不定参数传递，参数的类型为 `DialOption`，不定参数是指函数传入的参数个数为不定数量，可以不传，也可以为多个。

写一个不定参数传递的方法也很简单，看看下面这个方法 1 + 2 + 3 = 6。

```go
func Add(a int, args ...int) (result int) {
result += a
for _, arg := range args {
result += arg
}
return
}

fmt.Println(Add(1, 2, 3)) // 6
```

> 其实平时我们用的 `fmt.Println()`、`fmt.Sprintf()` 都属于不定参数的传递。

不定参数的几种形式：

```go
// 1. 收集到切片
func sum(nums ...int) int {
s := 0
for _, n := range nums {
s += n
}
return s
}

// 2. 已有切片用 ... 展开传
nums := []int{1, 2, 3}
sum(nums...)

// 3. 不定参数必须放在参数列表最后
// ❌ func f(a ...int, b int)  编译错误

// 4. 匿名函数也能接收不定参数
f := func(prefix string, args ...interface{}) {
fmt.Println(prefix, args...)
}
```

### 二、With 方法的作用

`WithInsecure()`、`WithBlock()` 类似于这样的 With 方法，其实作用就是 **修改 `dialOptions` 结构体的配置**
，之所以这样写我个人认为是面向对象的思想， **当配置项调整的时候调用方无需修改**。

### 场景

咱们模拟一个场景，使用 `不定参数` 和 `WithXXX` 这样的写法，写个 Demo。

比如我们要做一个从附近找朋友的功能，配置项有：性别、年龄、身高、体重、爱好。

我们要找性别为女性，年龄为 30 岁，身高为 160cm，体重为 55kg，爱好为爬山的人，希望是这样的调用方式：

```go
friends, err := friend.Find("附近的人",
friend.WithSex(1),
friend.WithAge(30),
friend.WithHeight(160),
friend.WithWeight(55),
friend.WithHobby("爬山"))
```

### 代码实现

> ⚠️ **先指出这个 Demo 的一处反模式（A16）**：下面代码用 `sync.Pool` 池化了一个只有 5 个字段的 `option` 结构体（约 48 字节）。
> **这是不值得的**：
>
> - 分配成本近乎为零，池化省不下什么；
> - 每次 `WithSex(1)` 仍会为返回的 **闭包**分配一次（闭包捕获了参数），这是池化省不掉的；
> - 净收益接近 0，代价却真实存在：必须维护 `reset()`， **漏一个字段就是脏数据串号**（上一个请求的 hobby 泄漏给下一个请求）。
>
> 正确做法见本节末尾的「对照实验」。`sync.Pool` 的正确使用场景是 **分配代价高的对象**（`bytes.Buffer`、大 `[]byte`、编解码器），详见第
> 08 篇四。

```go
// s10_option/friend/option.go
package friend

import (
	"sync"
)

// ⚠️ 反模式示例：为 48 字节的小对象维护 Pool，得不偿失（见上方说明）
var (
	cache = &sync.Pool{
		New: func() interface{} {
			return &option{sex: 0}
		},
	}
)

type Option func(*option)

type option struct {
	sex    int
	age    int
	height int
	weight int
	hobby  string
}

func (o *option) reset() {
	o.sex = 0
	o.age = 0
	o.height = 0
	o.weight = 0
	o.hobby = ""
}

func getOption() *option {
	return cache.Get().(*option)
}

func releaseOption(opt *option) {
	opt.reset()
	cache.Put(opt)
}

// WithSex setup sex, 1=female 2=male
func WithSex(sex int) Option {
	return func(opt *option) {
		opt.sex = sex
	}
}

// WithAge setup age
func WithAge(age int) Option {
	return func(opt *option) {
		opt.age = age
	}
}

// WithHeight setup height
func WithHeight(height int) Option {
	return func(opt *option) {
		opt.height = height
	}
}

// WithWeight setup weight
func WithWeight(weight int) Option {
	return func(opt *option) {
		opt.weight = weight
	}
}

// WithHobby setup hobby
func WithHobby(hobby string) Option {
	return func(opt *option) {
		opt.hobby = hobby
	}
}
```

```go
// s10_option/friend/friend.go
package friend

import (
	"fmt"
)

func Find(where string, options ...Option) (string, error) {
	friend := fmt.Sprintf("从 %s 找朋友\n", where)

	opt := getOption()
	defer func() {
		releaseOption(opt) // 用完立刻归还
	}()

	for _, f := range options {
		f(opt) // 逐个应用配置
	}

	if opt.sex == 1 {
		sex := "性别：女性"
		friend += fmt.Sprintf("%s\n", sex)
	}
	if opt.sex == 2 {
		sex := "性别：男性"
		friend += fmt.Sprintf("%s\n", sex)
	}

	if opt.age != 0 {
		age := fmt.Sprintf("年龄：%d岁", opt.age)
		friend += fmt.Sprintf("%s\n", age)
	}

	if opt.height != 0 {
		height := fmt.Sprintf("身高：%dcm", opt.height)
		friend += fmt.Sprintf("%s\n", height)
	}

	if opt.weight != 0 {
		weight := fmt.Sprintf("体重：%dkg", opt.weight)
		friend += fmt.Sprintf("%s\n", weight)
	}

	if opt.hobby != "" {
		hobby := fmt.Sprintf("爱好：%s", opt.hobby)
		friend += fmt.Sprintf("%s\n", hobby)
	}

	return friend, nil
}
```

```go
// s10_option/main.go
package main

import (
	"demo/friend"
	"fmt"
)

func main() {
	friends, err := friend.Find("附近的人",
		friend.WithSex(1),
		friend.WithAge(30),
		friend.WithHeight(160),
		friend.WithWeight(55),
		friend.WithHobby("爬山"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(friends)
}
```

### 输出

```text
从 附近的人 找朋友
性别：女性
年龄：30岁
身高：160cm
体重：55kg
爱好：爬山
```

### 这个 Demo 涉及的知识点汇总

| 知识点                               | 出现位置                                | 详见                                              |
|--------------------------------------|-----------------------------------------|---------------------------------------------------|
| 不定参数 `...Option`                 | `Find(where string, options ...Option)` | 本节一                                            |
| 函数类型 `type Option func(*option)` | `option.go`                             | 本节二                                            |
| 闭包捕获 option 指针                 | `WithSex` 等函数返回的闭包              | 第 05 篇三                                        |
| `defer` 归还资源                     | `defer releaseOption(opt)`              | 第 05 篇四                                        |
| `sync.Pool` 对象复用                 | `cache.Get()` / `cache.Put()`           | 第 08 篇四（**注意：此处属反模式，见 A16 说明**） |
| 私有结构体 + 构造函数                | `type option struct` / `New`            | 本节一                                            |
| 面向对象封装                         | 只有 `WithXxx` 对外暴露                 | 本节一                                            |

> 这一个小 Demo 串起了本教程 40% 的知识点，值得反复读。

### Options 模式的适用场景与代价

**适用**：

| 场景                 | 说明                     |
|----------------------|--------------------------|
| 配置项多且会持续增加 | 新增选项不破坏已有调用   |
| 多数配置用默认值     | 只传需要改的，其余走零值 |
| 未来可能扩展为链式   | `opt.WithX().WithY()`    |

**代价**：

- 参数校验被推迟到 **运行时**，编译期检查失效
- 需要维护 `reset()`，否则对象池会残留脏数据
- 大量选项会让签名很长，可读性下降

> **不要过度设计**：配置项少于 3 个、或必须全部指定时， **直白的结构体参数是更好的选择**：
> ```go
> type FindOptions struct {
> 	Sex    int
> 	Age    int
> 	Height int
> 	Weight int
> 	Hobby  string
> }
>
> func Find(where string, opts FindOptions) (string, error) {
> 	// 参数校验在函数内集中处理，编译器能检查字段名
> }
> ```
> Options 模式的真正价值在于 **向后兼容**——项目已发布、调用方遍布多个服务时，才值得为此付出复杂度。

## 本篇要点

1. Go 接口是 **隐式实现**，方法集匹配即实现，不需要写 `implements`。
2. `var _ I = (*T)(nil)` 做 **编译期断言**，接口变更时立刻报错。
3. 结构体定义成私有 + 构造函数返回接口 = 封装，隐藏实现细节。
4. 接口可以组合（内嵌），一个类型可实现多个接口。
5. `opts ...Option` + `type Option func(*option)` 就是 **Options 模式**，与 `grpc.NewClient` 同源（原 `grpc.Dial` 已废弃）。
6. **对照实验（A16 修正）**：三种写法按推荐顺序——① 默认选 **结构体参数**（保留编译期检查、无池化负担）；② 需要向后兼容时用
   **Options 模式**；③ `sync.Pool` 池化 option 属 **反模式**（对象太小，收益趋近 0，却有 `reset()` 脏数据风险），Pool
   应留给分配代价高的对象（见第 08 篇四）。
7. **不要滥用 `interface{}`**，它会触发逃逸并丢失类型安全。
8. Options 模式有真实代价（失去编译期检查、需维护 reset）， **配置项少时用结构体参数更好**。
