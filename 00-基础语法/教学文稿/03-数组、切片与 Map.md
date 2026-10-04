# 03 数组、切片与 Map

> 学完本篇你将能够：区分数组与切片的本质差异、熟练完成切片的截取追加删除、掌握 Map 的六种声明方式与删除操作、避开 nil map 的写入陷阱。

## 概述

本文合并了原三篇教程：数组、切片、Map。三者是 Go 中最常用的三种集合类型，放在一起来讲，因为它们之间存在演进关系：

```mermaid
graph LR
    A["数组 Array<br/>定长 · 值类型"] -->|"加长度字段<br/>解引用指针"| B["切片 Slice<br/>变长 · 含指针的值类型"]
    B -->|"底层是哈希表"| C["映射 Map<br/>键值对 · 无序"]

    style A fill:#1f4e79,color:#ffffff
    style B fill:#2d6a2d,color:#ffffff
    style C fill:#7a4a1f,color:#ffffff
```

---

## 一、数组 Array

数组是一个由**固定长度**的特定类型元素组成的序列，一个数组可以由零个或多个元素组成，一旦声明了，数组的长度就固定了，不能动态变化。

> `len()` 和 `cap()` 返回结果**始终一样**。

### 声明数组

```go
// s03_array/main.go
package main

import "fmt"

func main() {
	// 一维数组
	var arr1 [5]int
	fmt.Println(arr1) // [0 0 0 0 0]

	var arr2 = [5]int{1, 2, 3, 4, 5}
	arr3 := [5]int{1, 2, 3, 4, 5}
	arr4 := [...]int{1, 2, 3, 4, 5, 6} // [...] 自动推断长度
	arr5 := [5]int{0: 3, 1: 5, 4: 6}    // 下标初始化，未指定的取零值

	// 二维数组
	var arr6 = [3][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {3, 4, 5, 6, 7}}
	arr7 := [3][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {3, 4, 5, 6, 7}}
	arr8 := [...][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {0: 3, 1: 5, 4: 6}}

	fmt.Println(arr2, arr3, arr4, arr5)
	fmt.Println(arr6, arr7, arr8)
}
```

要点：

- `[5]int` 中的 `5` 是**类型的一部分**，不是长度字段
- `[...]int{...}` 让编译器自己数元素个数
- `[5]int{0: 3, 1: 5, 4: 6}` 下标初始化，未指定的元素取零值 → `[3 5 0 0 6]`
- 二维数组的元素类型是 `[5]int`，可以继续嵌套

### 注意事项

#### 一、数组不可动态变化

```go
var arr1 = [5]int{1, 2, 3, 4, 5}
arr1[5] = 6
// 报错：invalid array index 5 (out of bounds for 5-element array)
```

#### 二、数组是值类型，函数传参时复制整个数组

```go
func modifyArr(a [5]int) {
	a[1] = 20
}

func modifyArrPtr(a *[5]int) {
	a[1] = 20
}

func main() {
	var arr = [5]int{1, 2, 3, 4, 5}

	modifyArr(arr)     // 传值：函数内改动不影响外部
	fmt.Println(arr)   // [1 2 3 4 5]

	modifyArrPtr(&arr) // 传地址：改动生效
	fmt.Println(arr)   // [1 20 3 4 5]
}
```

> 传递大数组会有明显的内存拷贝开销，这是实际开发中**几乎总是用切片替代数组**的原因。

#### 三、类型必须完全一致才能赋值

```go
var arr  = [5]int{1, 2, 3, 4, 5}
var arr1 [5]int = arr //  OK
var arr2 [6]int = arr //  报错：cannot use arr (type [5]int) as type [6]int in assignment
```

> 即使元素类型相同，长度不同也是**不同的类型**，不能赋值。

---

## 二、切片 Slice

切片是一种**动态数组**，比数组操作灵活，长度不是固定的，可以进行追加和删除。

> `len()` 和 `cap()` 返回结果**可相同也可不同**。

### 切片与数组的对比

| | 数组 | 切片 |
|---|---|---|
| 长度 | 声明时固定 | 可变 |
| 类型 | 纯值类型 | 含指针的值类型（`{ptr, len, cap}`） |
| 函数传参 | 复制整个数组 | 复制切片头（指向同一底层数组） |
| `append` | 不支持 | 支持，自动扩容 |
| `nil` | 不适用 | 空切片可为 nil |
| 内存开销 | 大 | 小（结构体只有指针 + len + cap） |

> 切片的三要素：**底层数组指针 ptr + 长度 len + 容量 cap**。`s[i]` 实际是 `*(ptr + i*size)`。

### 声明切片

```go
// s04_slice/main.go
package main

import "fmt"

func main() {
	var sli1 []int        // nil 切片
	fmt.Printf("len=%d cap=%d slice=%v nil=%v\n", len(sli1), cap(sli1), sli1, sli1 == nil)

	sli2 := []int{}       // 空切片（非 nil）
	sli3 := []int{1, 2, 3, 4, 5} // 字面量
	sli4 := make([]int, 5)       // 长度 5，容量 5
	sli5 := make([]int, 5, 9)    // 长度 5，容量 9（预留空间）
	sli6 := sli3[1:3]            // 从已有切片切出

	fmt.Println(sli2, sli3, sli4, sli5, sli6)
}
```

> **nil 切片 vs 空切片**：两者都能 `append`、都能 `range`（循环 0 次）、`len()` 都是 0。区别在于：
> - `sli1 == nil` 为 `true`，`sli2 == nil` 为 `false`
> - JSON 序列化时，nil 切片输出 `null`，空切片输出 `[]`
> - nil 切片可以读可以 append，但**不能直接赋值元素**：`sli1[0] = 1` 会 panic

### 截取切片

```go
sli := []int{1, 2, 3, 4, 5, 6}

fmt.Println(sli[1])    // 2         索引
fmt.Println(sli[:])    // [1 2 3 4 5 6]  全部
fmt.Println(sli[1:])   // [2 3 4 5 6]    左闭右开
fmt.Println(sli[:4])   // [1 2 3 4]
fmt.Println(sli[0:3])  // [1 2 3]
fmt.Println(sli[0:3:4]) // [1 2 3]，容量限制为 4
```

**三索引切片 `a[low:high:max]`** 中 `max` 是**容量上限**。它解决了「切片扩容时意外写坏别人数据」的问题：

```go
s := make([]int, 3, 10)
s[0], s[1], s[2] = 1, 2, 3

sub := s[0:3:3] // 容量限制为 3
sub = append(sub, 99)
fmt.Println(s)   // [1 2 3 0 0 0 0 0 0 0] —— s 未被污染
fmt.Println(sub) // [1 2 3 99]
```

> 如果写成 `s[0:3]`（容量仍为 10），`append(sub, 99)` 会直接写进 `s[3]`，污染原切片。

### 追加切片

```go
sli := []int{4, 5, 6} // len=3 cap=3
sli = append(sli, 7)  // len=4 cap=6  扩容，容量变为 2 倍
sli = append(sli, 8)  // len=5 cap=6  容量够，不扩容
sli = append(sli, 9)  // len=6 cap=6
sli = append(sli, 10) // len=7 cap=12 再次扩容
```

> **append 扩容时 cap 会翻倍**。（Go 1.18+ 的具体策略是容量 256 以内翻倍、以上按约 1.25 倍增长，因此并非严格翻倍。）
>
> 扩容会**分配新数组并复制旧元素**，所以已知最终长度时优先预分配：
> ```go
> sli := make([]int, 0, 100) // 一次分配，后续 append 不再扩容
> ```
> 详见第 09 篇「逃逸分析场景 03」。

### 删除切片

Go 没有内置 delete 函数，需自己实现：

```go
sli := []int{1, 2, 3, 4, 5, 6, 7, 8}

// 删除尾部 2 个
sli = sli[:len(sli)-2]

// 删除开头 2 个
sli = sli[2:]

// 删除中间 2 个（下标 3、4）
sli = append(sli[:3], sli[3+2:]...)
```

封装成通用函数：

```go
// 删除下标为 i 的元素
func deleteElem(s []int, i int) []int {
	return append(s[:i], s[i+1:]...)
}

// 保留满足条件的元素
func filter(s []int, keep func(int) bool) []int {
	res := s[:0] // 复用底层数组，零分配（第 09 篇）
	for _, v := range s {
		if keep(v) {
			res = append(res, v)
		}
	}
	return res
}
```

> ⚠️ `append(s[:i], s[i+1:]...)` 复用底层数组，**如果原切片还有其他引用者（子切片），会污染它们的数据**。安全做法是显式复制：
> ```go
> res := make([]int, 0, len(s)-1)
> res = append(res, s[:i]...)
> res = append(res, s[i+1:]...)
> ```

### 复制

```go
src := []int{1, 2, 3}

// 目标切片必须已分配
dst := make([]int, len(src))
copy(dst, src) // 内置函数，复制元素，返回复制的个数

// 也可以把切片截成数组来复制（编译期会优化）
var arr [3]int
copy(arr[:], src)
```

---

## 三、映射 Map

Map 是**无序的 key-value** 数据结构。

> **key 必须是可比较类型（comparable）**：数值、字符串、布尔、指针、通道、接口，以及全部字段均可比较的结构体。
> **slice、map、func 不能做 key**（编译期报错）；含有这些类型字段的结构体也不能做 key。
> **value 才是任意类型**。所有 key 必须同一数据类型，所有 value 必须同一数据类型，key 和 value 的类型可以不相同。
>
> ```go
> m := map[string]int{}
> // m[[]int{1}] = 1  // ❌ 编译错误：invalid map key type []int
>
> var k any = []int{1}
> m2 := map[any]int{}
> m2[k] = 1         // ✅ 能编译（interface 可比较），但运行时 panic:
>                    // hash of unhashable type []int
> ```
>
> 实际开发中 key 最常用的是 `string` 和 `int`，结构体做 key 时注意不要包含 slice/map 字段。

### 声明 Map

```go
// s06_map/main.go
package main

import "fmt"

func main() {
	// 六种等价写法
	var p1 map[int]string
	p1 = make(map[int]string)
	p1[1] = "Tom"

	var p2 map[int]string = map[int]string{}
	var p3 map[int]string = make(map[int]string)
	p4 := map[int]string{}
	p5 := make(map[int]string)
	p6 := map[int]string{
		1: "Tom",
	}

	_ = p1
	_ = p2
	_ = p3
	_ = p4
	_ = p5
	_ = p6
}
```

> **nil map 的陷阱**：
> ```go
> var m map[string]int
> fmt.Println(m["a"])      // 0      读取安全
> delete(m, "a")           //        删除安全
> m["a"] = 1               // panic: assignment to entry in nil map
> ```
> 作为函数返回值或结构体字段时，通常要判断：
> ```go
> if m == nil {
> 	m = make(map[string]int)
> }
> ```

### 嵌套 Map 与 JSON 序列化

```go
res := make(map[string]interface{})
res["code"] = 200
res["msg"] = "success"
res["data"] = map[string]interface{}{
	"username": "Tom",
	"age":      "30",
	"hobby":    []string{"读书", "爬山"},
}
fmt.Println("map data :", res)

// 序列化
jsons, errs := json.Marshal(res)
if errs != nil {
	fmt.Println("json marshal error:", errs)
}
fmt.Println("--- map to json ---")
fmt.Println("json data :", string(jsons))

// 反序列化
res2 := make(map[string]interface{})
errs = json.Unmarshal(jsons, &res2)
if errs != nil {
	fmt.Println("json unmarshal error:", errs)
}
fmt.Println("--- json to map ---")
fmt.Println("map data :", res2)
```

> ⚠️ 这里的 `res2` 中所有数字都会变成 `float64`，这是第 04 篇要讲的经典坑。

### 编辑和删除

```go
person := map[int]string{
	1: "Tom",
	2: "Aaron",
	3: "John",
}
fmt.Println("data :", person)

person[2] = "Jack"  // 修改
person[4] = "Kevin" // 新增
fmt.Println("data :", person)

delete(person, 2)   // 删除
delete(person, 99)  // 删除不存在的 key 不会报错
fmt.Println("data :", person)
```

### 判断 key 是否存在

```go
val, ok := person[1]
if ok {
	fmt.Println("存在，值为", val)
}

// 忽略 ok
val2 := person[1]
fmt.Println(val2) // 不存在时为零值 ""
```

> 与之配套，`val == ""` **不能**用来判断 key 是否存在——key 存在但值恰好为空字符串时会误判。

### 遍历

```go
for k, v := range person {
	fmt.Printf("person[%d]: %s\n", k, v)
}
```

> ⚠️ **Map 遍历顺序是随机的**。Go 官方有意随机化以防止开发者依赖遍历顺序。需要稳定输出时，先把 key 收集起来排序：
> ```go
> keys := make([]int, 0, len(person))
> for k := range person {
> 	keys = append(keys, k)
> }
> sort.Ints(keys)
> for _, k := range keys {
> 	fmt.Println(k, person[k])
> }
> ```
> 这正是第 06 篇「生成签名」函数里必须 `sort.Strings(key)` 的原因。

### 常见用途

Map 适合表达「**键的集合是动态的**」的映射关系：

- 用字符串字段过滤结构体列表
- 缓存 / 字典 / 计数器
- 第三方接口的动态响应结构（详见第 04 篇）

**选型规则**：字段固定 → struct；字段动态 → map；`map[string]interface{}` 只用于**结构确实未知**的透传场景（性能代价见第 09 篇「两个开发注意点」）。

## 本篇要点

1. 数组定长、是纯值类型；切片是**含指针的值类型**（`{ptr, len, cap}` 结构体），变长且共享底层数组。**实际开发优先用切片**。
2. 切片 = 指针 + len + cap；`cap` 在数组中恒等于 `len`。
3. 三索引切片 `a[low:high:max]` 用容量隔离避免数据污染。
4. `append` 扩容会重新分配并复制，已知长度时用 `make([]T, 0, n)` 预分配。
5. Go 没有 delete 函数，删除靠 `append(s[:i], s[i+1:]...)` 组合。
6. nil map **可读可 delete 但不可写**。
7. Map 遍历无序，需要稳定顺序就排序 key。
8. `val, ok := m[k]` 判断存在性，不要用 `val == 零值`。
9. 字段固定用 struct，字段动态才用 map。
