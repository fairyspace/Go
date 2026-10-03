# 06 ｜ 函数、闭包与 defer

> **课时**：3 学时（135 分钟）｜**前置**：02、04
> **对应讲义**：`../教学文稿/06-函数、闭包与 defer.md`｜**对应示例**：`s08_func/` `s08_sign/` `s08_defer/`
> **本篇勘误**：A13、B11、C5（测试视角）、C7（术语）

---

## 1. 教学目标

### 1.1 知识目标
1. 能写出规范的函数签名：多返回值、命名返回值、不定参数、`error` 放最后。
2. 能准确说出 **Go 只有值传递**，并解释"传指针""切片/map 共享底层数据"分别意味着什么。
3. 能说出 `defer` 的四条规则：LIFO、参数在**声明时**求值、`return` 非原子、只对**当前 goroutine** 有效。

### 1.2 能力目标
1. 能在"值 / 指针 / 返回值"三种参数方案间做出正确选择。
2. 能用闭包写出计数器、中间件、资源归还三类模式。
3. 能徒手推演 defer 经典题的输出顺序。

### 1.3 素养目标
1. **错误处理惯例**：`error` 作最后一个返回值，成功返回 `nil`，调用方立即检查。
2. **资源释放紧跟获取**：`f, err := os.Open(...)` 的下一行就是 `defer f.Close()`。
3. **从本篇起，每个作业都要带 `_test.go`**（勘误 C5）。

## 2. 重点与难点

| | 内容 | 处理方式 |
|---|---|---|
| **重点** | 值传递语义 + 何时传指针 | 三行对比代码 |
| **重点** | defer 的 LIFO 与参数求值时机 | 经典题贯穿全课 |
| **难点** | `return` 非原子（命名返回值才能被 defer 改） | t1~t4 四段对比 |
| **难点** | 闭包捕获**变量**而非值 | 现场改外层变量看输出 |
| **难点** | defer 只对当前 goroutine 有效 | 现场让程序崩一次 |

## 3. 课前准备
投影；把 defer 经典题写在黑板上**先不公布答案**，全课围绕它推进。

## 4. 教学流程（135 分钟）

| 时间 | 环节 | 内容 |
|---|---|---|
| 00–08 | 导入 | 板书 defer 经典题，请学员猜输出（不公布） |
| 08–30 | 讲授 | 函数定义、多返回值、命名返回值、错误处理惯例 |
| 30–50 | 讲授 | 值传递 vs 传指针；切片/map 的特殊性（A13） |
| 50–70 | 讲授+演示 | 闭包：捕获变量、计数器、中间件 |
| 70–80 | 休息 | |
| 80–110 | 讲授+演示 | defer 五个坑 + 循环内 defer |
| 110–125 | 揭晓 | 回到经典题，逐行推演 |
| 125–135 | 练习+小结 | §7 + §5.6 |

## 5. 讲授要点（含板书）

### 5.1 板书一：函数签名规范

```go
func Divide(a, b float64) (float64, error) {   // error 永远在最后
	if b == 0 {
		return 0, errors.New("除数不能为 0")   // 或 fmt.Errorf("除数为 0: %w", ErrInvalidArg)
	}
	return a / b, nil
}

// 调用方：立即检查
res, err := Divide(10, 0)
if err != nil {
	return fmt.Errorf("计算失败: %w", err)   // 包装时保留错误链
}
```

**命名返回值**
```go
func Sum(a, b int) (sum int) {
	sum = a + b
	return            // 裸 return，等价于 return sum
}
```
两个正当用途：① 配合 `defer` 修改返回值（§5.5）；② 文档可读性。**不要**为了省写字滥用裸 return。

**不定参数**
```go
func Sum(nums ...int) int {          // 函数内 nums 的类型是 []int
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}
Sum(1, 2, 3)
nums := []int{1, 2, 3}
Sum(nums...)                          // 已有切片用 ... 展开
// ❌ func f(a ...int, b int)          不定参数必须在最后
```

### 5.2 板书二：Go 只有值传递（勘误 C7）

```go
type Result struct{ Code int }

func byValue(r Result)  { r.Code = 500 }   // 改副本
func byPtr(r *Result)   { r.Code = 500 }   // 改本体
```

**速查表（修正 A13）**

| 需求 | 正确做法 |
|---|---|
| 修改基本类型 | `func f(p *int)` + `f(&n)` |
| 修改 struct 字段 | `func f(p *T)` + `f(&v)` |
| 修改切片/map 的**已有元素** | 直接传值即可：`s[i] = v`、`m[k] = v`（共享底层数据） |
| 改变切片的**长度** | **返回新切片** `s = f(s)`；或传 `*[]T` |
| 只读大对象、避免拷贝 | `func f(p *T)`（但小对象传指针可能因逃逸更慢，见 09 篇） |
| 明确不想被修改 | `func f(v T)` 传值 |

```go
// ❌ 调用方看不到
func tryAppend(s []int) { s = append(s, 1, 2, 3) }

// ✅ 标准做法
func appendAll(s []int, vs ...int) []int { return append(s, vs...) }
s = appendAll(s, 1, 2, 3)
```

> 一句话记住：**传进去的都是副本**。切片副本里含指针，所以能改元素；但副本里的 `len` 改了，调用方的 `len` 不变。

### 5.3 板书三：闭包

**闭包 = 函数 + 它捕获的变量（变量本身，不是值的快照）**
```go
a := 1
add := func() { fmt.Println(a) }
a = 2
add()          // 输出 2，不是 1
```

**三个实用模式**
```go
// ① 计数器
func newCounter() func() int {
	i := 0
	return func() int { i++; return i }
}

// ② 中间件 / 装饰器
func withLog(next func(string)) func(string) {
	return func(name string) {
		log.Println("进入", name)
		next(name)
		log.Println("退出", name)
	}
}

// ③ 资源归还（07 篇 Options 模式会用）
opt := getOption()
defer releaseOption(opt)
```

**循环变量（Go 版本差异，必讲）**
```go
for i := 1; i <= 3; i++ {
	go func() { fmt.Println(i) }()
}
// Go ≤1.21：可能全打印 3（共享同一个 i）
// Go ≥1.22：每轮 i 是新变量，打印 1 2 3（顺序不定）
// 老代码里的 `i := i` 就是为了在旧语义下复制一份
```

### 5.4 板书四：defer 的规则

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
fmt.Println("main")
// 输出：main / 3 / 2 / 1     ← LIFO
```

**规则一：LIFO（后进先出）**——最先声明的最后执行。

**规则二：参数在 defer 语句执行时求值，函数体在执行时才求值**
```go
a, b := 1, 2

defer fmt.Println(a + b)              // 求值 3，稍后打印 3
defer func() { fmt.Println(a + b) }() // 闭包，执行时读 a → 4
defer func(x, y int) { fmt.Println(x + y) }(a, b) // 显式传参 → 3

a = 2
// 输出顺序（LIFO）：3 → 4 → 3
```

**规则三：`return` 不是原子操作**
```text
return expr  展开为：
  1. 把 expr 赋给返回值变量
  2. 执行所有 defer
  3. 真正返回
```
```go
func t1() int        { a := 1; defer func(){ a++ }(); return a }        // 1
func t2() (a int)    { defer func(){ a++ }(); return 1 }                // 2  ← 命名返回值可被 defer 改
func t3() (b int)    { a := 1; defer func(){ a++ }(); return 1 }        // 1  ← defer 改的不是返回值
func t4() (a int)    { defer func(x int){ x++ }(a); return 1 }          // 1  ← 传参是副本
```

**规则四：`os.Exit` 不执行 defer**
```go
func main() {
	defer fmt.Println("不会打印")
	os.Exit(0)
}
```
> 生产代码里尽量不用 `os.Exit`；必须用时，把清理逻辑放进**显式调用**的函数。

**规则五：defer（含 `recover`）只对当前 goroutine 有效**
```go
func GoA() {
	defer func() { recover() }()   // 捕获不到 GoB 的 panic
	go GoB()
}
func GoB() { panic("boom") }      // 整个进程崩溃
```
> 结论：**每个 goroutine 入口都要自己 `recover`**（08 篇给出 `safeGo` 模板）。

### 5.5 板书五：defer 的正确用法与循环陷阱

```go
// 资源释放：紧跟获取
f, err := os.Open(name)
if err != nil {
	return err
}
defer f.Close()          // ← 就写在这里，不要挪到函数末尾

// 锁：保证解锁
mu.Lock()
defer mu.Unlock()
```

**循环内 defer 会在函数结束时才统一执行**
```go
// ❌ 10 次循环累积 10 个 defer，锁直到函数结束才释放
for _, v := range items {
	mu.Lock()
	defer mu.Unlock()
	handle(v)
}

// ✅ 包一层函数，每轮立即执行
for _, v := range items {
	func() {
		mu.Lock()
		defer mu.Unlock()
		handle(v)
	}()
}
```

**性能**：Go 1.14 起有 open-coded defer，开销接近普通调用；但**每秒百万级的热路径**仍应避免。

### 5.6 小结（口述七句）
1. `error` 放最后，调用方立即检查，包装用 `%w`。
2. Go **只有值传递**；改长度要返回新切片。
3. 闭包捕获**变量本身**。
4. defer 是 LIFO。
5. defer 的**参数**在声明时求值，**函数体**在执行时求值。
6. `return` 非原子，只有**命名返回值**能被 defer 修改。
7. `os.Exit` 不执行 defer；defer 不跨 goroutine；循环内 defer 要包函数。

## 6. 演示脚本：defer 经典题（全课主线）

```go
package main

import "fmt"

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

**推演过程（要求学员能复述）**
```go
// 等价改写：defer 的参数会立即求值
tmp1 := calc("B", 1, 2)   // 立即执行 → 打印 "B 1 2 3"，返回 3
defer calc("A", 1, 3)     // 登记
x = 3
tmp2 := calc("D", 3, 2)   // 立即执行 → 打印 "D 3 2 5"，返回 5
defer calc("C", 3, 5)     // 登记
y = 4                     // 与已登记的参数无关
// 函数结束，LIFO 执行：
// calc("C", 3, 5) → "C 3 5 8"
// calc("A", 1, 3) → "A 1 3 4"
```

**输出**
```text
B 1 2 3
D 3 2 5
C 3 5 8
A 1 3 4
```

> 两个考点：① 嵌套调用的参数**立即求值**（B、D 先打印）；② defer 登记的是**值**，后续 `x = 3`、`y = 4` 影响不到已登记的 A、C。

## 7. 课堂练习（含参考答案）

**练习 1** 输出什么？
```go
func f() (result int) {
	defer func() { result *= 2 }()
	result = 10
	return
}
```
> **答**：`20`。`return`（裸）时 `result` 已是 10，defer 把命名返回值改成 20。

**练习 2** 输出什么？
```go
func f() int {
	result := 10
	defer func() { result *= 2 }()
	return result
}
```
> **答**：`10`。返回值是**匿名**的，`return result` 先把 10 复制给返回值，defer 改的是局部变量 `result`，改不到返回值。

**练习 3** 下面代码有什么问题？给出修正。
```go
func readAll(paths []string) error {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		process(f)
	}
	return nil
}
```
> **答**：循环内 defer，所有文件句柄要等函数返回才关闭；`paths` 很长时会耗尽文件描述符。修正：把循环体抽成函数（或包一层匿名函数），让 `defer` 每轮执行。

**练习 4** 实现 `func Retry(times int, fn func() error) error`：失败重试 `times` 次，全部失败则返回最后一个错误（用 `errors.Join` 收集全部错误更好）。写单测覆盖：首次成功、末次成功、全部失败。
> **答**：
> ```go
> func Retry(times int, fn func() error) error {
> 	var errs []error
> 	for i := 0; i < times; i++ {
> 		if err := fn(); err != nil {
> 			errs = append(errs, err)
> 			continue
> 		}
> 		return nil
> 	}
> 	return errors.Join(errs...)   // Go 1.20+
> }
> ```

## 8. 勘误与易错预警

| 编号 | 讲义原说法 | 正确说法 |
|---|---|---|
| **A13** | 修改切片/Map 不用传指针 | 改**元素**不用；改**长度**必须返回新切片 |
| **C7** | "引用传递" | Go 只有值传递；说"传指针"或"共享底层数据" |
| **B11** | 示例用 `str = str + fmt.Sprintf(...)`，正文却反对 `+` | 用 `strings.Builder`；签名前缀统一为 `k=v&k=v` |
| — | 学员把 `defer` 写在函数末尾 | 必须**紧跟资源获取**，否则中间的 `return` 漏掉清理 |
| — | 学员在 goroutine 外写 `recover` 想兜住内部 panic | 无效（规则五）；每个 goroutine 自己 recover |
| — | 学员用 `defer fmt.Println(err)` 记录错误 | defer 里读到的是**声明时**的 `err`（通常是 nil）；要用闭包 |

## 9. 课后作业（分层，均需 `_test.go`）

- **基础**：实现 `SafeDivide(a, b float64) (float64, error)`，除零返回可识别的哨兵错误 `ErrDivideByZero`；单测用 `errors.Is` 断言。
- **进阶**：实现签名函数（对齐 `s08_sign`，但修正 B11）：
  ```go
  func CreateSign(params map[string]any, secret string) string
  ```
  要求：key 排序、`k=v` 用 `&` 连接（**用 `strings.Builder`**）、`MD5(MD5(body) + MD5(secret))`。单测断言：同一入参多次调用结果一致；参数顺序打乱后结果不变；secret 不同则签名不同。
- **挑战**：实现 `func WithTimeout(d time.Duration, fn func() error) error`（提示：用 goroutine + channel + `select` + `time.After`），并写单测覆盖"按时完成"与"超时"两条路径。思考题：超时后那个 goroutine 去哪了？（答：泄漏——为 08 篇的 context 埋点。）

## 10. 考核点（可观测）

- [ ] 能徒手推演 §6 经典题并说出两个考点。
- [ ] 能说出 t1~t4 四个函数的返回值及原因。
- [ ] 能解释"函数内 append 调用方看不到"并给出正确签名。
- [ ] 能写出"资源获取后紧跟 defer"的标准代码。
- [ ] 能说出 defer 不跨 goroutine，并写出 `safeGo` 模板。
- [ ] 进阶作业的 `go test ./...` 通过，且签名对参数顺序不敏感。

## 11. 延伸阅读
- Go 博客：*Defer, Panic, and Recover*
- Go 规范：Defer statements、Return statements
- Go 1.14 发布说明：open-coded defer
- 下一课预告：接口的隐式实现、编译期断言 `var _ I = (*T)(nil)`，以及 `grpc.NewClient(target, opts...)` 背后的 Options 模式。
