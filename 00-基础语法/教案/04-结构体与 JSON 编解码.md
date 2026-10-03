# 04 ｜ 结构体与 JSON 编解码

> **课时**：3 学时（135 分钟）｜**前置**：02、03
> **对应讲义**：`../教学文稿/04-结构体与 JSON.md`｜**对应示例**：`s05_struct/` `s12_json/`
> **本篇勘误**：A3、A4、A19

---

## 1. 教学目标

### 1.1 知识目标
1. 能说出 struct 的**值语义**：赋值与传参都复制全部字段。
2. 能解释值接收者与指针接收者的差别，并说出"同一类型接收者风格必须统一"的理由。
3. 能说清 `json.Unmarshal` 到 `map[string]interface{}` 时数字**一律变成 `float64`**，并区分**显示问题**与**精度问题**。

### 1.2 能力目标
1. 用 struct tag 控制字段名、省略零值、排除字段。
2. 针对三种对接场景选择正确解析策略（强类型 / 弱类型 / 完全动态）。
3. 用 `json.Number`（`UseNumber`）保住大整数的原始字面量。

### 1.3 素养目标
1. **传输层字段（ID、金额、序列号）永远不要用 `map[string]interface{}` 接收**——这是线上事故高发点。
2. 结构明确就定义 struct；不要为了"灵活"牺牲类型安全。

## 2. 重点与难点

| | 内容 | 处理方式 |
|---|---|---|
| **重点** | 值语义 + 何时传指针 | 现场演示"改了没生效" |
| **重点** | JSON tag 四个常用选项 | 表格 + 现场打印 Marshal 结果 |
| **难点** | `float64` 的显示 vs 精度（A4） | 用两个边界值对比：1.76e15 与 9.007e15+1 |
| **难点** | 三种解析策略的选型 | 决策表 + 场景演练 |
| **难点** | `json.Number` 参与运算必须显式转换 | 现场制造编译错误 |

## 3. 课前准备
投影；准备一段第三方接口的真实响应 JSON（含 19 位 ID 与金额）作为贯穿案例。

## 4. 教学流程（135 分钟）

| 时间 | 环节 | 内容 |
|---|---|---|
| 00–10 | 导入 | 展示一个订单号变成 `1.234567890123e+18` 的线上事故截图 |
| 10–35 | 讲授 | struct 声明、嵌套、值语义、比较 |
| 35–50 | 讲授 | 接收者风格：值 vs 指针 |
| 50–70 | 讲授+演示 | JSON tag、Marshal/Unmarshal、Encoder/Decoder |
| 70–80 | 休息 | |
| 80–105 | 讲授 | 三种解析策略 + 科学计数法/精度双问题 |
| 105–120 | 演示 | `UseNumber` + `json.Number` 转换 |
| 120–135 | 练习+小结 | §7 四题 + §5.6 |

## 5. 讲授要点（含板书）

### 5.1 板书一：结构体

```go
type Person struct {
	Name string
	Age  int
}

var p1 Person                 // 零值 {Name:"" Age:0}
fmt.Printf("%+v\n", Person{}) // {Name: Age:0}

p2 := Person{Name: "Burke", Age: 31}  // 推荐：字段名初始化
p3 := Person{Name: "Aaron"}           // 未写的字段取零值（Age=0）
p4 := struct {                        // 匿名结构体：临时数据、测试用例
	Name string
	Age  int
}{Name: "匿名", Age: 33}
```

**为什么推荐字段名初始化**：字段顺序变化时不会静默错位；`go vet` 也能检查未导出字段的跨包初始化问题。

**嵌套与比较**
```go
type Address struct{ City, Zip string }
type User struct {
	Name    string
	Addr    Address    // 值嵌套：复制一份
	Ptr     *Address   // 指针嵌套：共享同一份
}

a := Address{"杭州", "310000"}
b := Address{"杭州", "310000"}
fmt.Println(a == b)   // true —— 所有字段都可比较时，struct 可比较
```
> 含 slice / map / func 字段的 struct **不能用 `==`**（编译错误），需要 `reflect.DeepEqual` 或自己写 `Equal` 方法。

### 5.2 板书二：值语义与接收者

```go
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func setByValue(r Result)  { r.Code = 500 }   // 改副本，调用方看不到
func setByPtr(r *Result)   { r.Code = 500 }   // 改本体
```

**接收者选择规则（按顺序判断）**
1. 需要修改接收者 → **指针**。
2. struct 较大（> 64 字节）或含 `sync.Mutex` 等不可复制字段 → **指针**（含 Mutex 时是**必须**）。
3. 接收者是 map / slice / chan / func → 值即可（它们本身就含指针）。
4. 其余情况 → 值。
5. **同一类型风格必须统一**：混用会让方法集规则（`T` 的方法集不含 `*T` 的指针方法）导致接口实现失败——这是最隐蔽的 bug 之一。

> ⚠️ 传指针**不一定更快**：小对象传指针可能触发逃逸到堆（见 09 篇）。默认按上面规则选，别凭感觉。

### 5.3 板书三：JSON tag

| tag | 作用 | 注意 |
|---|---|---|
| `json:"name"` | 字段名映射 | 大小写敏感匹配有回退规则，但**别依赖** |
| `json:"name,omitempty"` | 零值时省略 | 零值定义：`false`/`0`/nil 指针/nil 接口/空数组切片 map/空串 |
| `json:"-"` | **完全不参与 JSON（双向跳过）** | 勘误 **A3** |
| `json:"-,"` | 字段名就是 `-` | 特殊写法 |
| `json:"id,string"` | 数值以 JSON 字符串形式收发 | 反序列化时 JSON 里必须是**带引号的字符串**，否则报错 |

```go
type Optional struct {
	Name  string `json:"name"`
	Age   int    `json:"age,omitempty"`
	Skip  string `json:"-"`
	Dash  string `json:"-,"`
}
o, _ := json.Marshal(Optional{Name: "Tom", Dash: "x"})
// {"name":"Tom","-":"x"}   ← Age 被省略，Skip 消失，Dash 以 "-" 为键
```

**序列化三种姿势**
```go
b, err := json.Marshal(v)                      // 一次性
b, err = json.MarshalIndent(v, "", "  ")       // 带缩进（排错用）
err = json.NewEncoder(w).Encode(v)             // 流式：直接写 io.Writer，省一次 []byte
```

**反序列化两种姿势**
```go
err := json.Unmarshal(data, &v)                            // 一次性
err = json.NewDecoder(r).Decode(&v)                        // 流式
dec := json.NewDecoder(r); dec.UseNumber(); dec.Decode(&v) // 保住数字字面量
dec.DisallowUnknownFields()                                // 严格模式：多余字段报错
```

> ⚠️ nil 切片序列化成 `null`。要输出 `[]`，初始化成空切片（`make([]T, 0)`），而不是靠 `omitempty`——`omitempty` 是**整个字段消失**，语义不同。

### 5.4 板书四：三种解析策略（决策表）

| 场景 | 策略 | 代码要点 |
|---|---|---|
| **结构明确**（99% 的情况） | **强类型 struct** | 定义 struct + tag；性能最好、类型安全 |
| 字段名稳定但**类型不稳定**（`"200"` 有时是 int 有时是 string） | 弱类型转换 | `map[string]any` + 手写/`mapstructure.WeakDecode` |
| 字段**完全不固定**，只做透传/存储 | `map[string]any` + **`UseNumber()`** | 显式转换每一个数字 |
| 字段不固定但**层级固定** | `json.RawMessage` 延迟解析 | 先存原始字节，确定类型后再解 |

**手写弱类型转换（无需第三方库）**
```go
func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)   // 注意大数精度问题
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}
```

**内嵌结构体天然摊平**
```go
type Family struct{ LastName string }
type Location struct{ City string }
type Person struct {
	Family            // 匿名内嵌 → JSON 里 LastName 是顶层字段
	Location
	FirstName string
}
```
> `mapstructure` 需要写 `mapstructure:",squash"`；**`encoding/json` 默认就摊平**，不要写 `json:",squash"`（它不认识这个选项，勘误 B4）。

### 5.5 板书五：数字的两个问题（勘误 A4，必须讲透）

**问题一：显示成科学计数法（值没变）**
```go
var m map[string]any
json.Unmarshal([]byte(`{"number":1234567}`), &m)
fmt.Println(m)                    // map[number:1.234567e+06]
fmt.Printf("%v\n", m["number"])   // 1.234567e+06
n, _ := m["number"].(float64)
fmt.Printf("%.0f\n", n)           // 1234567  ← 值是对的，只是 %v 的显示方式
```
原因：`json.Unmarshal` 到 `any` 时数字一律存成 `float64`；`float64` 经 `%v` 输出时，有效位 ≥ 7 位就切换为科学计数法。**这是格式化行为，不是数据损坏。**

**问题二：真的丢精度（值变了）**
```
边界：2^53 = 9007199254740992（约 9.007e15，16 位）

1759482245000000  (1.76e15) < 2^53 → 精确表示，不丢精度
9007199254740993  (2^53+1)          → 变成 9007199254740992，丢 1
7300000000000000001 (19 位雪花 ID)   → 变成 7.3e+18，尾部全丢
```
> ⚠️ **讲义里的例子举错了**：`1759482245000000` 小于 2^53，能被 float64 精确表示，它只是**显示**成科学计数法。要演示精度丢失，必须用 **> 9007199254740992** 的值，例如 `9007199254740993` 或 19 位雪花 ID。

**三种解法**
```go
// 解法一：强转（治标）——精度已经丢了也救不回来
id := int64(m["id"].(float64))

// 解法二：定义 struct（治本，首选）
type Order struct {
	ID     int64   `json:"id"`      // 直接解成 int64，全程无 float64
	Amount float64 `json:"amount"`  // 金额更推荐用"分为单位的整数"或 shopspring/decimal
}

// 解法三：UseNumber（结构未知时）
dec := json.NewDecoder(bytes.NewReader(data))
dec.UseNumber()
dec.Decode(&m)
num := m["id"].(json.Number)        // 类型是 json.Number
s := num.String()                  // "7300000000000000001" 原始字面量，一位不丢
i, err := num.Int64()              // 转 int64
f, err := num.Float64()            // 转 float64
```
> `json.Number` 的底层是 `string`。它**不能直接参与算术**（`num + 1` 编译错误），必须先 `Int64()`/`Float64()`——这既是它的麻烦，也是它的安全。

### 5.6 小结（口述七句）
1. struct 是值语义，**要改就传指针**；含 `sync.Mutex` 必须指针接收者。
2. 同一类型接收者风格统一。
3. `json:"-"` 双向跳过；输出 `-` 用 `json:"-,"`。
4. 解析优先级：**struct > RawMessage > map + UseNumber**。
5. `any` 接数字一律变 `float64`：**≥7 位显示成科学计数法**（值没坏），**>2^53 才真丢精度**。
6. `json.Number` 保住字面量，但要显式转换才能运算。
7. **ID / 金额 / 序列号绝不用 `map[string]interface{}` 接**。

## 6. 演示脚本

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Secret  string `json:"-"`
}

func main() {
	// 1) tag 行为
	r := Result{Code: 200, Message: "ok", Secret: "不输出"}
	b, _ := json.Marshal(r)
	fmt.Println("1)", string(b))            // {"code":200,"msg":"ok"}

	// 2) 显示 vs 精度
	const small = `{"id":1759482245000000}`   // < 2^53
	const big = `{"id":7300000000000000001}`  // 19 位雪花 ID
	for _, s := range []string{small, big} {
		var m map[string]any
		json.Unmarshal([]byte(s), &m)
		fmt.Printf("2) 原始=%s  map=%v  取整=%d\n",
			s, m["id"], int64(m["id"].(float64)))
	}

	// 3) UseNumber 保住字面量
	dec := json.NewDecoder(bytes.NewReader([]byte(big)))
	dec.UseNumber()
	var m2 map[string]any
	dec.Decode(&m2)
	num := m2["id"].(json.Number)
	fmt.Println("3) 类型:", reflect.TypeOf(num), " 字面量:", num.String())
	i64, _ := num.Int64()
	fmt.Println("   Int64:", i64)             // 7300000000000000001 一位不丢

	// 4) 强类型解析（首选）
	type Order struct {
		ID int64 `json:"id"`
	}
	var o Order
	json.Unmarshal([]byte(big), &o)
	fmt.Println("4) struct 解析:", o.ID)
}
```

预期输出要点：第 2 段中 `small` 的 `map` 显示为 `1.759482245e+15` 但**取整仍是 1759482245000000**（未丢精度）；`big` 的取整结果尾部失真。第 3、4 段均完整保留 19 位。

## 7. 课堂练习（含参考答案）

**练习 1** 预测输出：
```go
type T struct {
	A int    `json:"a,omitempty"`
	B string `json:"-"`
	C string `json:"-,"`
}
b, _ := json.Marshal(T{C: "x"})
fmt.Println(string(b))
```
> **答**：`{"-":"x"}`。A 是零值被省略，B 双向跳过，C 的键名是 `-`。

**练习 2** 下面代码哪里错了？给出两种修法。
```go
var m map[string]any
json.Unmarshal([]byte(`{"id":7300000000000000001}`), &m)
id := m["id"].(int64)
```
> **答**：`m["id"]` 的动态类型是 `float64` 不是 `int64`，类型断言会 **panic**。修法一：`int64(m["id"].(float64))`（但大数已丢精度）；修法二：用 `UseNumber()` + `num.Int64()`，或直接定义 `struct{ ID int64 }`。

**练习 3** 写出"金额"字段的三种可选建模方式，并说明各自适用场景。
> **答**：① `int64` 存"分"（最稳，推荐用于计费）；② `shopspring/decimal`（需要精确十进制运算，如汇率、税费）；③ `float64`（仅用于展示型统计，不能用于对账）。**任何情况下都不要让金额经过 `map[string]any`。**

**练习 4** 为什么 `json:",string"` 有时会让反序列化失败？
> **答**：该选项要求 JSON 里的值是**带引号的字符串**（`{"id":"123"}`）。如果对端发的是 `{"id":123}`，`Unmarshal` 会报 `json: invalid use of ,string struct tag`。跨端联调时要与对端约定死格式。

## 8. 勘误与易错预警

| 编号 | 讲义原说法 | 正确说法 |
|---|---|---|
| **A3** | `json:"-"` 只用于反序列化 | **双向跳过**；要输出字面量 `-` 用 `json:"-,"` |
| **A4** | `1759482245000000` 会丢精度 | 它 < 2^53，**精确表示**；丢精度的门槛是 `> 9007199254740992` |
| **A19** | `p3 := Person{Name:"Aaron", Age:32} // 忽略 Age` | 注释与代码矛盾；演示部分初始化应写 `Person{Name:"Aaron"}` |
| **B4** | 结构体里写 `json:",squash"` | `encoding/json` 不认识 squash；匿名内嵌默认摊平 |
| — | 学员用 `omitempty` 想把 nil 切片变成 `[]` | `omitempty` 是**整个字段消失**；要 `[]` 请初始化空切片 |
| — | 学员忘记 `Unmarshal` 传指针 | 传值会得到 `InvalidUnmarshalError` |

## 9. 课后作业（分层）

- **基础**：定义 `User{Name, Email string; Age int; Password string}`，要求 `Password` 不出现在 JSON 中，`Age` 为 0 时省略；写单测断言 Marshal 结果字符串。
- **进阶**：实现 `func ParseResponse(data []byte) (*Order, error)`，能正确处理：ID 为 19 位整数、金额为字符串 `"12.34"`、`created_at` 为 RFC3339。要求**不使用 `map[string]any`**。
- **挑战**：实现一个 `type Int64String int64`，让它序列化成 JSON 字符串、反序列化时同时接受 `"123"` 与 `123`（实现 `MarshalJSON` / `UnmarshalJSON`），并写单测覆盖两种输入与非法输入。

## 10. 考核点（可观测）

- [ ] 能说出 struct 值语义，并解释何时必须用指针接收者。
- [ ] 能默写四个 JSON tag 选项的含义，特别是 `json:"-"` 与 `json:"-,"` 的区别。
- [ ] 能说出 `2^53` 这个边界，并解释"显示问题"与"精度问题"的区别。
- [ ] 能用 `UseNumber` 保住 19 位 ID，并说出 `json.Number` 的底层类型。
- [ ] 进阶作业的 `go test ./...` 通过。

## 11. 延伸阅读
- `pkg.go.dev/encoding/json`（Marshal / Unmarshal 的字段匹配规则）
- Go 1.25+ 的实验性 `encoding/json/v2` 与 `encoding/json/jsontext`（Go 1.27 起 v2 成为 `encoding/json` 的后端）
- 下一课预告：流程控制——`range` 的值是副本、`switch` 默认不穿透、以及 `goto` 为什么不该出现在业务代码里。
