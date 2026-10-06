# 04 结构体与 JSON

> 学完本篇你将能够：用 struct 描述复杂数据、掌握 JSON 序列化的三种解析方式、避开 `json.Unmarshal` 最经典的精度坑、知道何时该用
> struct 何时该用 map。

## 概述

本文合并了原三篇教程：结构体、解析 JSON 数据、json.Unmarshal 遇到的小坑。它们是一条完整链路： **结构体是载体 → JSON
是传输格式 → 解析是过程 → 踩坑与解法**。

---

## 一、结构体 Struct

结构体是将零个或多个任意类型的变量，组合在一起的 **聚合类型**数据类型，也可以看做是数据的集合。

### 声明结构体

```go
// s05_struct/main.go
package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func main() {
	var p1 Person // 取零值
	p1.Name = "Tom"
	p1.Age = 30
	fmt.Println("p1 =", p1)

	p2 := Person{Name: "Burke", Age: 31} // 字段名初始化
	p3 := Person{Name: "Aaron", Age: 32} // 忽略 Age
	fmt.Println("p2 =", p2, "p3 =", p3)

	// 匿名结构体：常用于临时数据、测试用例
	p4 := struct {
		Name string
		Age  int
	}{Name: "匿名", Age: 33}
	fmt.Println("p4 =", p4)
}
```

要点：

- 字段名 **首字母大写导出**（`Name`），小写私有（`name`）
- 字段可以任意类型，包括数组、切片、map、其他结构体
- 零值演示（须在赋值 **之前**打印）：`fmt.Printf("%+v\n", Person{})` 输出 `{Name: Age:0}`

### 嵌套与比较

```go
type Address struct {
City string
Zip  string
}

type User struct {
Name    string
Age     int
Address Address // 值嵌套（拷贝一份）
Ptr     *Address // 指针嵌套（共享同一份）
}

u := User{Name: "Tom", Address: Address{City: "杭州", Zip: "310000"}}
fmt.Println(u.Address.City) // 杭州
```

结构体可整体比较（要求所有字段可比较）：

```go
a := Address{City: "杭州", Zip: "310000"}
b := Address{City: "杭州", Zip: "310000"}
fmt.Println(a == b) // true
```

> 结构体含 slice、map、func 字段时 **不可用 `==` 比较**，需用 `reflect.DeepEqual(a, b)`。

### 值传递 vs 指针传递

结构体默认按值传递会 **拷贝整个结构体**。要修改外部变量，必须传指针：

```go
type Result struct {
Code    int    `json:"code"`
Message string `json:"msg"`
}

func setData(res *Result) { // 指针参数
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

> **与第 05 篇的关系**：这个「传指针才能修改」的规则，与 `defer` 的参数取值时机是同一个语言机制的两种表现。
>
> ⚠️ **传指针不一定更快**：小数据量传指针会导致对象逃逸到堆（见第 08 篇）。

### 指针接收者 vs 值接收者

```go
type Counter struct{ n int }

func (c *Counter) Add() { c.n++ }    // 指针接收者：修改生效
func (c Counter) Get() int { return c.n } // 值接收者：只读

var c Counter
c.Add() // Go 会自动取地址，等价于 (&c).Add()
fmt.Println(c.Get())
```

> **约定**：同一类型的接收者风格必须统一。struct 含 `sync.Mutex` 等不可复制字段时， **必须**用指针接收者。

---

## 二、JSON 序列化

### 结构体 tag

```go
type Result struct {
Code    int    `json:"code"`
Message string `json:"msg"`
}
```

常用 tag 选项：

| 选项                           | 作用                                                         |
|--------------------------------|--------------------------------------------------------------|
| `json:"name"`                  | 字段名映射                                                   |
| `json:"name,omitempty"`        | 为零值时省略该字段                                           |
| `json:"-"`                     | 该字段**完全不参与 JSON**（Marshal 与 Unmarshal 双向都跳过） |
| `json:"-,"`                    | 字段名就叫 `-`（带逗号才生效，否则被当作上面的忽略标记）     |
| `json:",string"`               | 数值以字符串形式输出                                         |
| `json:"name,string,omitempty"` | 组合使用                                                     |

> ⚠️ 常见误区：`json:"-"` 并不是「只用于反序列化」。官方文档明确：
> "if the field tag is `-`, the field is always omitted. Note that a field with name `-` can still be generated using
> the tag `- ,`"（即 `json:"-,"`）。
> 想做「序列化时输出、反序列化时忽略」（常见于密码字段），需要 **另定义一个不tag该字段的接收结构体**，或实现 `json.Unmarshaler`
> 接口。

> tag 是 **反射**机制：Marshal/Unmarshal 通过 `reflect` 读取 tag 来决定字段映射关系。

### 序列化与反序列化

```go
res := Result{Code: 200, Message: "success"}

// 序列化：结构体 → JSON 字符串
jsons, errs := json.Marshal(res)
if errs != nil {
fmt.Println("json marshal error:", errs)
}
fmt.Println("json data :", string(jsons)) // {"code":200,"msg":"success"}

// 反序列化：JSON 字符串 → 结构体
var res2 Result
errs = json.Unmarshal(jsons, &res2)
if errs != nil {
fmt.Println("json unmarshal error:", errs)
}
fmt.Println("res2 :", res2) // {200 success}
```

常用变体：

```go
// 格式化输出（排查问题用）
fmt.Println(string(jsons))

data, _ := json.MarshalIndent(res, "", "  ")
fmt.Println(string(data))

// 直接写流，省去一次 string 转换
json.NewEncoder(os.Stdout).Encode(res)
```

> ⚠️ 切片/数组 nil 时输出 `null`，需用 `omitempty` 规避。

---

## 三、解析 JSON 数据的三种情况

> 现实场景：对接第三方接口，响应结构可能不稳定。

### 情况一：结构明确 —— 强类型解析

请求手机归属地接口，返回如下：

```json
{
  "resultcode": "200",
  "reason": "Return Successd!",
  "result": {
    "province": "浙江",
    "city": "杭州",
    "areacode": "0571",
    "zip": "310000",
    "company": "中国移动",
    "card": ""
  }
}
```

**思路**：

1. 先将 json 转成 struct
2. 然后 `json.Unmarshal()` 即可

json 转 struct 手写太麻烦，推荐在线工具：https://mholt.github.io/json-to-go/
（左边贴 JSON，右边自动生成 struct）

```go
type MobileInfo struct {
Resultcode string `json:"resultcode"`
Reason     string `json:"reason"`
Result     struct {
Province string `json:"province"`
City     string `json:"city"`
Areacode string `json:"areacode"`
Zip      string `json:"zip"`
Company  string `json:"company"`
Card     string `json:"card"`
} `json:"result"`
}

func main() {
jsonStr := `
	{
		"resultcode": "200",
		"reason": "Return Successd!",
		"result": {
			"province": "浙江",
			"city": "杭州",
			"areacode": "0571",
			"zip": "310000",
			"company": "中国移动",
			"card": ""
		}
	}
	`

var mobile MobileInfo
err := json.Unmarshal([]byte(jsonStr), &mobile)
if err != nil {
fmt.Println(err.Error())
return
}
fmt.Println(mobile.Resultcode) // 200
fmt.Println(mobile.Reason) // Return Successd!
fmt.Println(mobile.Result.City) // 杭州
}
```

> 这就是 **最推荐的做法**：结构明确时永远用 struct，性能最好（第 08 篇会说明为什么）。

### 情况二：数据类型不确定 —— 弱类型解析

先定义一个 `string` 类型的 `resultcode`，JSON 却返回了 `int`：

```json
{
  "resultcode": 200
}
```

思路是去 github 找开源类库，使用 **mapstructure**：https://github.com/mitchellh/mapstructure

它有一个弱类型解析方法 `WeakDecode()`：

```go
type MobileInfo struct {
Resultcode string `mapstructure:"resultcode"`
}

func main() {
jsonStr := `{"resultcode": 200}`

// 第一步：先转成通用 map
var result map[string]interface{}
if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
fmt.Println(err.Error())
return
}

// 第二步：弱类型转换到 struct
var mobile MobileInfo
if err := mapstructure.WeakDecode(result, &mobile); err != nil {
fmt.Println(err.Error())
return
}

fmt.Println(mobile.Resultcode) // 200（int 自动转为 string）
}
```

`WeakDecode()` 会做类型强转（int → string、number → bool 等）。

> ⚠️ **用 mapstructure 时 struct tag 要写 `mapstructure` 而不是 `json`**：
> ```go
> Code int `mapstructure:"code" json:"code"`
> ```
> 两个都写上，可以同时服务两种场景。

### 情况三：result 中参数不固定 —— squash 嵌入

使用官方示例 `Example (EmbeddedStruct)`：

```go
type Family struct {
LastName string
}
type Location struct {
City string
}
type Person struct {
Family    `mapstructure:",squash"`
Location  `mapstructure:",squash"`
FirstName string
}

func main() {
input := map[string]interface{}{
"FirstName": "Mitchell",
"LastName":  "Hashimoto",
"City":      "San Francisco",
}

var result Person
if err := mapstructure.Decode(input, &result); err != nil {
panic(err)
}

fmt.Println(result.FirstName) // Mitchell
fmt.Println(result.LastName) // Hashimoto
fmt.Println(result.City) // San Francisco
}
```

> `,squash` 表示把内嵌结构体的字段 **摊平**到外层，这样 `LastName` 就是 `result.LastName` 而不是`result.Family.LastName`。
>
> 其他常用 tag：
> - `mapstructure:",remain"` —— 未被匹配到的字段统一收进该字段
> - `mapstructure:"name"` —— 重命名字段
> - `mapstructure:",omitempty"` —— 编码时省略零值

字段极不确定时，直接返回 `map[string]interface{}` 即可——但要注意下面第四节的坑。

---

## 四、json.Unmarshal 遇到的小坑：科学计数法

### 1. 问题现象描述

使用 `json.Unmarshal()` 反序列化时，出现了科学计数法：

```go
jsonStr := `{"number":1234567}`
result := make(map[string]interface{})
err := json.Unmarshal([]byte(jsonStr), &result)
if err != nil {
fmt.Println(err)
}
fmt.Println(result)

// 输出：map[number:1.234567e+06]
```

> **这个问题不是必现**，只有当数字的位数大于 6 位时，才会变成科学计数法。

### 2. 问题影响描述

当数据结构未知、使用 `map[string]interface{}` 来接收反序列化结果时，如果数字的位数大于 6 位，都会变成科学计数法，用到的地方都会受到影响。

典型受害场景： **订单号、雪花 ID、金额、时间戳（13 位毫秒时间戳受影响最严重）**。

### 3. 引起问题的原因

从 `encoding/json` 可以找到答案，看一下这段注释：

```go
// To unmarshal JSON into an interface value,
// Unmarshal stores one of these in the interface value:
//
//	bool, for JSON booleans
//	float64, for JSON numbers
//	string, for JSON strings
//	[]interface{}, for JSON arrays
//	map[string]interface{}, for JSON objects
//	nil for JSON null
```

是因为当 `JSON` 中存在一个比较大的数字时，它会被解析成 `float64` 类型，就有可能会出现科学计数法的形式。

> 这里要 **区分两个不同的问题**，成因不同、必须分开理解：
>
> **① 显示问题（值没变）**：`float64` 经 `%v` 输出时，有效位 ≥ 7 位就切换成科学计数法。`1234567` 显示为 `1.234567e+06`
> ，但数值本身是精确的——`float64(1234567)` 与 `1234567` 相等。
>
> **② 精度问题（值真的变了）**：`float64` 只有 53 位尾数，能 **精确表示**的整数范围是 `±2^53`（`9007199254740992`，约 9.0e15）。
> **整数绝对值超过 2^53 才会真正丢精度**。
>
> ```go
> // 可实机验证
> fmt.Println(float64(1234567) == 1234567)                 // true —— 只是显示成科学计数法
> fmt.Println(float64(1759482245000000))                   // 1.759482245e+15 —— 仍精确（< 2^53）
> fmt.Println(float64(7300000000000000001) == 7300000000000000001) // false —— 19 位雪花 ID，真丢精度
> ```
>
> 典型的精度受害者是 **19 位雪花 ID**（如 `7300000000000000001`）、超过 2^53 的纳秒级计算结果。两个问题的解法相同（别用
> `map[string]interface{}` 接 ID），但成因不同。

### 4. 问题的解决方案

#### 方案一：强制类型转换

```go
fmt.Println(int(result["number"].(float64))) // 1234567
```

> ⚠️ 只在确认数值在 `int` 范围内时安全，超大 ID 应转 `int64`：
> ```go
> fmt.Println(int64(result["number"].(float64)))
> ```
> 但这 **无法挽回已经丢失的精度**，只是把问题推迟。

#### 方案二：定义结构体接收（推荐）

尽量避免使用 `interface`，对 json 字符串结构定义结构体。快捷方法可使用在线工具：https://mholt.github.io/json-to-go/

```go
type Num struct {
Number int `json:"number"`
}

jsonStr := `{"number":1234567}`
var result Num
err := json.Unmarshal([]byte(jsonStr), &result)
if err != nil {
fmt.Println(err)
}
fmt.Println(result) // {1234567}
```

> 这是 **最推荐的方案**。它同时解决了科学计数法、精度丢失和类型安全问题。
> 完整代码见 `codes/教学/s12_json/main.go`。

#### 方案三：使用 `UseNumber()` 方法

```go
jsonStr := `{"number":1234567}`
result := make(map[string]interface{})
d := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
d.UseNumber() // 关键：让数字保持原始字符串形态
err := d.Decode(&result)
if err != nil {
fmt.Println(err)
}
fmt.Println(result) // map[number:1234567]
```

这时 **类型变成 `json.Number`**，注意一定要记住这一点：

```go
fmt.Println(fmt.Sprintf("type: %v", reflect.TypeOf(result["number"])))
// 输出：type: json.Number
```

通过代码可以看出 `json.Number` 其实就是字符串类型：

```go
// A Number represents a JSON number literal.
type Number string
```

如果转换其他类型，参考如下代码：

```go
// 转成 int64
numInt, _ := result["number"].(json.Number).Int64()
fmt.Println(fmt.Sprintf("value: %v, type: %v", numInt, reflect.TypeOf(numInt)))
// 输出：value: 1234567, type: int64

// 转成 string
numStr := result["number"].(json.Number).String()
fmt.Println(fmt.Sprintf("value: %v, type: %v", numStr, reflect.TypeOf(numStr)))
// 输出：value: 1234567, type: string
```

> ⚠️ 用 `UseNumber()` 后，所有数字都是 `json.Number`， **直接参与算术会编译报错**，必须先转 `Int64()` / `Float64()`
> 。这既是它的安全之处，也是它的烦人之处。

---

## 五、三种方案怎么选

| 场景                                     | 推荐方案                                                     |
|------------------------------------------|--------------------------------------------------------------|
| **结构明确**（最常见）                   | 方案二：定义 struct                                          |
| 结构不明确但字段名稳定，需要访问具体字段 | 情况二：`mapstructure.WeakDecode`                            |
| 结构完全不固定，只做透传/存储            | 方案三：`UseNumber()` + 显式转换                             |
| 结构不固定且字段少                       | 考虑用 `map[string]interface{}`，但**ID/金额字段要单独处理** |

> **最重要的原则**： **传输层字段（ID、金额、序列号）永远不要用 `map[string]interface{}` 接收。** 这是线上事故的高发点。

## 本篇要点

1. struct 是值类型， **要修改必须传指针**；含 `sync.Mutex` 的 struct 必须用指针接收者。
2. 同一类型的接收者风格必须统一（都值或都指针）。
3. `json` tag 通过 **反射**生效，字段名映射、omitempty、`-` 排除都靠它。
4. 解析策略： **明确用 struct** → **类型不定用 `WeakDecode`** → **字段不定用 `,squash`** → **完全不固定用 `map` +
   `UseNumber`**。
5. 用 mapstructure 时 tag 写 `mapstructure:"..."`。
6. **`json.Unmarshal` 到 `interface{}` 时数字一律变 `float64`**。有效位 ≥ 7 位就显示为科学计数法； **整数绝对值超过
   2^53（9007199254740992）才真正丢精度**。三种解法：强转 / 定义 struct / `UseNumber()`。
7. `UseNumber()` 后的数字是 `json.Number`（底层是 string），必须显式转 `Int64()` / `Float64()`。
8. 字段固定用 struct，字段动态才用 map（性能原因见第 08 篇）。
