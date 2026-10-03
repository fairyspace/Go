# 06 函数、闭包与 defer

> 学完本篇你将能够：写出规范函数签名、在值传递与引用传递间正确选择、看懂闭包的引用语义、避开 defer 的五个经典坑、写出带缓存复用的签名函数。

## 概述

本篇合并了原「函数」和「defer 函数」两篇教程。它们本质是同一个主题：**Go 的函数调用语义**——参数何时求值、结果何时返回、副作用何时发生。

---

## 一、函数定义

```go
func function_name(input1 type1, input2 type2) (type1, type2) {
	// 函数体
	// 返回多个值
	return value1, value2
}
```

- 函数用 `func` 声明
- 函数可以有一个或多个参数，需要有参数类型，用 `,` 分割
- 函数可以有一个或多个返回值，需要有返回值类型，用 `,` 分割
- 函数的参数是可选的，返回值也是可选的

### 命名返回值与裸 return

```go
func getSum(a, b int) (sum int) {
	sum = a + b
	return // 等价于 return sum
}
```

> 命名返回值在 Go 中主要是为了配合 `defer` 修改返回值（第 4 节坑 3），以及文档可读性。

### 错误处理惯例

```go
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为 0")
	}
	return a / b, nil
}

res, err := divide(10, 0)
if err != nil {
	fmt.Println("error:", err)
	return
}
fmt.Println(res)
```

> Go 惯例把 `error` 作为**最后一个返回值**，成功时返回 `nil`。

### 常见工具函数

```go
// MD5
func MD5(str string) string {
	s := md5.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

// 获取当前时间字符串
func getTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 获取当前时间戳
func getTimeInt() int64 {
	return time.Now().Unix()
}
```

> Go 的时间格式布局是**参考时间** `2006-01-02 15:04:05`，不是 Java 那套 `yyyy-MM-dd`。详见第 10 篇。

### 实战：生成签名

```go
// s08_sign/main.go
package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
)

func main() {
	params := map[string]interface{}{
		"name": "Tom",
		"pwd":  "123456",
		"age":  30,
	}
	fmt.Printf("sign : %s\n", createSign(params))
}

// MD5 方法
func MD5(str string) string {
	s := md5.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

// 生成签名
func createSign(params map[string]interface{}) string {
	// 1. 取出所有 key
	var key []string
	for k := range params {
		key = append(key, k)
	}

	// 2. 排序（map 遍历无序，必须排序）
	sort.Strings(key)

	// 3. 拼接字符串
	var str = ""
	for i := 0; i < len(key); i++ {
		if i == 0 {
			str = fmt.Sprintf("%v=%v", key[i], params[key[i]])
		} else {
			str = str + fmt.Sprintf("&xl_%v=%v", key[i], params[key[i]])
		}
	}

	// 4. 自定义密钥
	var secret = "123456789"

	// 5. 双重 MD5
	return MD5(MD5(str) + MD5(secret))
}
```

> 实际项目可在此基础上增加**时间戳**（满足时效性）和**随机 nonce**（满足唯一性），并从配置文件读取 secret。完整签名四要素见第 09 篇。
>
> ⚠️ 拼接大量字符串时不要用 `+`，原因见第 09 篇；MD5/AES 等各算法的性能数据也见第 09 篇。

---

## 二、值传递与引用传递

**值传递**：传递参数时，将参数**复制一份**传递到函数中，对参数进行调整后，不影响参数值。

**Go 语言默认是值传递。**

**引用传递**：传递参数时，将参数的**地址**传递到函数中，对参数进行调整后，影响参数值。

```go
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func setData(res *Result) {
	res.Code = 500
	res.Message = "fail"
}

func toJson(res *Result) {
	jsons, errs := json.Marshal(res)
	if errs != nil {
		fmt.Println("json marshal error:", errs)
	}
	fmt.Println("json data :", string(jsons))
}

func main() {
	res := Result{Code: 200, Message: "success"}
	toJson(&res)

	setData(&res) // 传指针，修改生效
	toJson(&res)
}
```

> 结构体较大时按值传递会有明显拷贝开销；修改外部变量必须用指针。但**小数据量传指针反而可能更慢**（会触发逃逸，见第 09 篇）。

**常见指针写法速查**：

| 需求 | 写法 |
|---|---|
| 修改结构体字段 | `func f(p *T)` + 调用 `f(&v)` |
| 修改切片/Map | 不用传指针（切片含指针，Map 是引用类型） |
| 修改基本类型 | `func f(p *int)` + 调用 `f(&n)` |
| 只读大对象，避免拷贝 | `func f(p *T)` 但不修改 |
| 避免修改外部变量 | `func f(v T)` 传值 |

---

## 三、闭包

**闭包 = 函数 + 它捕获的外部变量**。当函数内部引用了外层的局部变量时，就形成了闭包。

```go
func main() {
	var a = 1
	var b = 2

	// 闭包捕获的是变量本身，不是值
	add := func() {
		fmt.Println("a+b =", a+b)
	}

	a = 3
	add() // 输出 4，而不是 3
}
```

> **结论：闭包获取变量相当于引用传递，而非值传递。**

这带来两个实用特性：

**1. 闭包 + 循环变量（Go 1.22 之前必须手动复制）**

```go
// Go 1.22 之前：所有闭包共享同一个 i，全部输出 3
for i := 1; i <= 3; i++ {
	go func() { fmt.Println(i) }()
}
```

**Go 1.22 起**，`for` 循环的每轮迭代都有独立变量，此问题自动解决。旧代码需要：

```go
for i := 1; i <= 3; i++ {
	i := i // 显式复制
	go func() { fmt.Println(i) }()
}
```

**2. 闭包 + defer 构成资源归还模式**（第 7 篇的 Options 模式就用到了）

```go
opt := getOption()
defer func() {
	releaseOption(opt) // 用完立刻归还
}()
```

**3. 闭包 + 循环变量做并发任务生成器**

```go
func newCounter() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}
```

---

## 四、defer

defer 函数大家肯定都用过，它在**声明时不会立刻去执行**，而是在函数 **return 后**去执行。

主要应用场景：异常处理、记录日志、清理数据、释放资源等等。

> 本节不是分享 defer 的应用场景，而是分享**使用 defer 需要注意的点**。咱们先从一道题开始。

### 经典题

```go
func calc(index string, a, b int) int {
	ret := a + b
	fmt.Println(index, a, b, ret)
	return ret
}

func main() {
	x := 1
	y := 2
	defer calc("A", x, calc("B", x, y))
	x = 3
	defer calc("C", x, calc("D", x, y))
	y = 4
}
```

输出什么？答案在最后一节，先往下看。

### 坑 1：执行顺序是 LIFO（后进先出）

```go
func main() {
	defer fmt.Println("1")
	defer fmt.Println("2")
	defer fmt.Println("3")

	fmt.Println("main")
}
```

输出：

```text
main
3
2
1
```

> **结论**：defer 函数定义的顺序与实际执行的顺序是相反的，也就是**最先声明的最后才执行**。

### 坑 2：闭包 vs 传参

```go
var a = 1
var b = 2

defer fmt.Println(a + b) // 参数在 defer 时就确定了
a = 2
fmt.Println("main")
// 输出：main 3
```

```go
var a = 1
var b = 2

defer func() {
	fmt.Println(a + b) // 函数体内变量在执行时才确定
}()
a = 2
fmt.Println("main")
// 输出：main 4
```

```go
var a = 1
var b = 2

defer func(a int, b int) {
	fmt.Println(a + b) // 显式传参 → 值复制
}(a, b)
a = 2
fmt.Println("main")
// 输出：main 3
```

> **结论**：
> - `defer fmt.Println(a+b)` → 参数值在 **defer 定义时**就确定了
> - `defer func() { ... }()` → 函数内变量值在**函数运行时**才确定（闭包引用传递）
> - `defer func(a, b int) { ... }(a, b)` → 传参是**值复制**
>
> 这条规则来自第 02 篇：`:=` 声明的变量只在当前作用域内可见，而 defer 语句会立即求值它的参数。

### 坑 3：Return 不是原子操作

```go
func t1() int {
	a := 1
	defer func() { a++ }()
	return a
}
// 输出：1
```

```go
func t2() (a int) {
	defer func() { a++ }()
	return 1
}
// 输出：2
```

```go
func t3() (b int) {
	a := 1
	defer func() { a++ }() // 修改的不是返回值 b
	return 1
}
// 输出：1
```

```go
func t4() (a int) {
	defer func(a int) { a++ }(a) // 值传递，改的是副本
	return 1
}
// 输出：1
```

> **结论**：`return` 不是原子操作。
>
> `return a` 实际展开为三步：
> ```text
> 1. ret = a         // 赋值给返回值变量
> 2. 执行所有 defer  // defer 可以修改它
> 3. return           // 真正返回
> ```
>
> 只有**命名返回值**才能被 defer 修改（t2）。t1 之所以是 1，是因为 `return a` 中的 `a` 是普通局部变量；t3 的返回值叫 `b`，defer 改的 `a` 与它无关。

### 坑 4：os.Exit 不执行 defer

```go
func main() {
	defer fmt.Println("1")
	fmt.Println("main")
	os.Exit(0)
}
// 输出：main
```

> **结论**：当 `os.Exit()` 方法退出程序时，defer 不会被执行。
>
> 这意味着 defer 里的资源释放、日志落盘、事务回滚**全部失效**。生产代码中 `os.Exit` 应尽量避免，或把所有清理逻辑放进显式调用的函数里。

### 坑 5：defer 只对当前协程有效

```go
func main() {
	GoA()
	time.Sleep(1 * time.Second)
	fmt.Println("main")
}

func GoA() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic:" + fmt.Sprintf("%s", err))
		}
	}()

	go GoB()
}

func GoB() {
	panic("error")
}
```

> `GoB()` 的 panic **捕获不到**。
>
> **结论**：defer（含 `recover`）只对**当前协程**有效。**每个 goroutine 都要自己写 `recover`**，主 goroutine 的 recover 管不到子 goroutine。
>
> 这个问题怎么解？咱们下回再说（见第 08 篇的并发安全 checklist）。

### 答案解析

先列出答案：

```text
B 1 2 3
D 3 2 5
C 3 5 8
A 1 3 4
```

其实上面那道题，可以拆解为：

```go
func main() {
	x := 1
	y := 2
	tmp1 := calc("B", x, y) // 立即执行
	defer calc("A", x, tmp1)
	x = 3
	tmp2 := calc("D", x, y) // 立即执行
	defer calc("C", x, tmp2)
	y = 4
}
```

所以顺序就是 **B → D → C → A**：

- 执行到 `tmp1` 时，输出：`B 1 2 3`
- 执行到 `tmp2` 时，输出：`D 3 2 5`
- 根据 defer 执行顺序原则（先声明的后执行），下一个该执行 C 了。又因为**传参是值赋值**，所以 C 拿不到 `y = 4`，输出：`C 3 5 8`
- 然后执行 A。A 同时拿不到 `x = 3` 和 `y = 4`，输出：`A 1 3 4`

到这，基本上 defer 就清楚了，大家可以根据自己的理解去记忆。

---

## 五、defer 的正确用法

```go
// 资源释放的标准写法
func handle(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close() // 立即注册，后续任何 return 都会执行

	// 业务逻辑...
	return nil
}

// 配合对象池（第 09 篇）
stu := student.New("tom", 30)
defer student.Release(stu)

// 配合锁
mu.Lock()
defer mu.Unlock() // 保证解锁一定执行
```

### 循环内 defer 的坑

`for` 循环里的 defer 会在**函数结束时**才全部执行，而不是每轮迭代执行：

```go
// ❌ 错误：10 次循环会累积 10 个锁，直到函数结束才释放
for _, v := range items {
	mu.Lock()
	defer mu.Unlock()
	handle(v)
}
```

需要每轮释放时，把它包成函数：

```go
// ✅ 正确
for _, v := range items {
	func() {
		mu.Lock()
		defer mu.Unlock()
		handle(v)
	}()
}
```

或把循环体抽成独立函数：

```go
func process(v Item) {
	mu.Lock()
	defer mu.Unlock()
	handle(v)
}

for _, v := range items {
	process(v)
}
```

> **defer 的性能**：Go 1.14 起 defer 实现了 **open-coded defer**，开销已降低到接近普通函数调用（原文基于 Go 1.12 时代，defer 开销较大，现代 Go 已不构成瓶颈）。但仍**不要在每秒百万次调用的热路径中使用 defer**。

## 本篇要点

1. Go 默认值传递；修改结构体/大对象要传指针，但小对象传指针可能因逃逸更慢。
2. 闭包捕获的是**变量本身**而非值。
3. defer **LIFO** 执行。
4. defer 的**参数在定义时求值**（值复制）；defer **函数体里的变量在执行时求值**。
5. `return` 不是原子操作，只有**命名返回值**能被 defer 修改。
6. `os.Exit()` 不执行 defer——生产代码要格外小心。
7. **defer 只对当前协程有效，每个 goroutine 都要自己 `recover()`。**
8. 循环内 defer 要包成函数才会每轮执行。
9. Go 1.14+ 的 open-coded defer 开销已很小，但热路径仍应避免。
