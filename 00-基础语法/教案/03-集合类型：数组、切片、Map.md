# 03 ｜ 集合类型：数组、切片、Map

> **课时**：3 学时（135 分钟，建议拆成 2 次课）｜**前置**：02
> **对应讲义**：`../教学文稿/03-数组、切片与 Map.md`｜**对应示例**：`s03_array/` `s04_slice/` `s06_map/`
> **本篇勘误**：A12、A13（部分）、A20、C7（术语）

---

## 1. 教学目标

### 1.1 知识目标
1. 能从**内存布局**说出数组与切片的差别：数组是 N 个元素的连续块；切片是 `{ptr, len, cap}` 三元组。
2. 能解释"共享底层数组"带来的数据污染，并说出三索引切片 `s[low:high:max]` 的作用。
3. 能说出 map 的三条硬约束：**key 必须可比较**、遍历无序、nil map 可读不可写。

### 1.2 能力目标
1. 熟练完成切片的声明、截取、追加、删除、复制。
2. 会用 `make([]T, 0, n)` 预分配，并解释它省掉了什么。
3. 会写"按 key 排序后遍历 map"以获得稳定输出。

### 1.3 素养目标
1. **默认用切片，不用数组**；数组只在长度是类型语义的一部分时使用（如 `[16]byte` 的 MD5 摘要）。
2. 写函数时想清楚"我要不要改调用方的长度"——这决定了返回值签名。

## 2. 重点与难点

| | 内容 | 处理方式 |
|---|---|---|
| **重点** | 切片三要素；append 扩容；共享底层数组 | 画图 + 打印 `ptr/len/cap` 观察 |
| **重点** | map 的存在性判断 `v, ok := m[k]` | 对比 `v == 零值` 的错误写法 |
| **难点** | 三索引切片与容量隔离 | 现场污染一次数据，再修好 |
| **难点** | 函数内 `append` 调用方看不到（A13） | 现场演示"改了但没生效" |
| **难点** | map key 的 comparable 约束（A12） | 现场制造编译错误 + 运行时 panic |

## 3. 课前准备
投影；准备一张切片内存图（可手绘：底层数组 + ptr/len/cap 三个框）。

## 4. 教学流程（135 分钟）

| 时间 | 环节 | 内容 |
|---|---|---|
| 00–10 | 导入 | 「为什么 `append` 之后，我传进去的切片没变？」——真实 bug 现场 |
| 10–30 | 讲授 | 数组：定长、值语义、长度是类型的一部分 |
| 30–60 | 讲授+演示 | 切片三要素、截取、三索引、append 扩容 |
| 60–75 | 演示 | 共享底层数组的数据污染 + 三种删除 |
| 75–80 | 休息 | |
| 80–105 | 讲授 | Map：声明、nil 陷阱、key 约束、无序遍历 |
| 105–120 | 演示 | 稳定遍历；`v, ok` 判断存在性 |
| 120–135 | 练习+小结 | §7 四题 + §5.5 |

## 5. 讲授要点（含板书）

### 5.1 板书一：数组 vs 切片

```
数组 [5]int           切片 []int（slice header，24 字节）
┌───┬───┬───┬───┬───┐   ┌─────┬─────┬─────┐
│ 0 │ 1 │ 2 │ 3 │ 4 │   │ ptr │ len │ cap │──┐
└───┴───┴───┴───┴───┘   └─────┴─────┴─────┘  │
  长度 5 是类型的一部分                        ▼
  [5]int 和 [6]int 是不同类型             底层数组
```

| | 数组 | 切片 |
|---|---|---|
| 长度 | 声明时固定，是**类型的一部分** | 可变 |
| 赋值/传参 | 复制**全部元素** | 复制 **header**（24 字节），共享底层数组 |
| `len` 与 `cap` | 恒相等 | 可不等 |
| 实际使用 | 极少（`[16]byte` 摘要、定长协议字段） | **默认选择** |

```go
arr := [5]int{1, 2, 3, 4, 5}
arr2 := [...]int{1, 2, 3}          // 编译器数长度
arr3 := [5]int{0: 3, 4: 6}         // 下标初始化 → [3 0 0 0 6]
var other [6]int = arr             // ❌ 类型不同，编译错误
```

### 5.2 板书二：切片的六种声明

```go
var s1 []int              // nil 切片：ptr=nil len=0 cap=0
s2 := []int{}             // 空切片：非 nil，len=0 cap=0
s3 := []int{1, 2, 3}      // 字面量
s4 := make([]int, 5)      // len=5 cap=5，元素全零值
s5 := make([]int, 5, 9)   // len=5 cap=9（预留空间）
s6 := s3[1:3]             // 从已有切片切出，共享底层数组
```

**nil 切片 vs 空切片**：`append`、`range`、`len` 行为**完全一致**；差别只有两点——
1. `s1 == nil` 为 `true`；
2. `json.Marshal` 时 nil 切片输出 `null`，空切片输出 `[]`（接口约定要 `[]` 时，必须初始化成空切片）。

> 写 `func Find() []User` 时，**返回 `[]User{}` 而不是 nil**，前端才不会拿到 `null`。

### 5.3 板书三：截取与三索引

```go
s := []int{1, 2, 3, 4, 5, 6}
s[1]        // 2
s[:]        // [1 2 3 4 5 6]
s[1:]       // [2 3 4 5 6]      左闭右开
s[:4]       // [1 2 3 4]
s[0:3:4]    // len=3, cap=4-0=4  ← 第三个数是"容量右边界"
```

**为什么需要三索引**：限制 `cap` 可以让 `append` 提前触发扩容，从而**不去写别人的内存**。
```go
s := make([]int, 3, 10)
s[0], s[1], s[2] = 1, 2, 3

sub := s[0:3:3]          // cap 被限制为 3
sub = append(sub, 99)    // 触发扩容 → 分配新数组
fmt.Println(s)           // [1 2 3 0 0 0 0 0 0 0]  s 干净

sub2 := s[0:3]           // cap 仍是 10
sub2 = append(sub2, 99)  // 直接写进 s[3]
fmt.Println(s)           // [1 2 3 99 0 ...]       s 被污染
```

### 5.4 板书四：append、预分配、删除

```go
s := []int{4, 5, 6}   // len=3 cap=3
s = append(s, 7)      // 扩容：分配新数组 + 复制 + 追加
```
- **必须写 `s = append(s, x)`**：append 可能返回一个指向新数组的 header。
- 扩容策略（Go 1.18+）：小切片约翻倍，超过约 256 元素后按 ~1.25 倍平滑增长。**不要依赖具体倍数**，只记住"会分配 + 复制"。
- **已知长度就预分配**：`s := make([]int, 0, n)`，把 n 次可能的扩容变成 1 次分配。

```go
// 删除：Go 没有内置的 slice delete（内置 delete 只用于 map）
s = s[:len(s)-1]                    // 删尾
s = s[1:]                           // 删头
s = append(s[:i], s[i+1:]...)       // 删中间（会移动后续元素，且污染共享者）

// Go 1.21+ 标准库：推荐
import "slices"
s = slices.Delete(s, i, i+1)        // 删 [i, i+1)
slices.Sort(s)
if slices.Contains(s, x) { ... }
```

> ⚠️ `append(s[:i], s[i+1:]...)` 复用底层数组，**若原切片还有别的引用者（子切片），它们的数据会被改坏**。安全写法：
> ```go
> res := make([]int, 0, len(s)-1)
> res = append(res, s[:i]...)
> res = append(res, s[i+1:]...)
> ```

### 5.5 板书五：Map

```go
var m1 map[string]int             // nil map
m2 := map[string]int{}            // 空 map
m3 := make(map[string]int)        // 空 map
m4 := make(map[string]int, 100)   // 预分配 100 个桶的容量提示
m5 := map[string]int{"a": 1}
```

**三条硬约束**
1. **key 必须是可比较类型（comparable）**。`slice`、`map`、`func` 不能做 key（编译错误）；含这些字段的结构体也不行。
   ```go
   m := map[any]int{}
   var k any = []int{1}
   m[k] = 1        // ⚠️ 运行时 panic: hash of unhashable type []int
   ```
2. **遍历顺序随机**（运行时故意随机化，防止开发者依赖顺序）。
3. **nil map 可读可删不可写**：
   ```go
   var m map[string]int
   _ = m["a"]        // ✅ 零值
   delete(m, "a")    // ✅ 空操作
   m["a"] = 1        // ❌ panic: assignment to entry in nil map
   ```

**存在性判断**
```go
v, ok := m[k]
if !ok { /* 不存在 */ }

// ❌ 错误写法：值恰好是零值时会误判
if m[k] == "" { /* 这可能只是值本来就是空串 */ }
```

**稳定遍历**
```go
keys := make([]string, 0, len(m))
for k := range m {
	keys = append(keys, k)
}
slices.Sort(keys)          // Go 1.21 前用 sort.Strings(keys)
for _, k := range keys {
	fmt.Println(k, m[k])
}
```

```go
// Go 1.23+ 的一行版：maps.Keys 返回迭代器，slices.Sorted 直接收集并排序
import "maps"
for _, k := range slices.Sorted(maps.Keys(m)) {
	fmt.Println(k, m[k])
}
```
> 这正是 06 篇「生成签名」必须先排序 key 的原因：**签名要求可复现**。

### 5.6 术语纠偏（勘误 C7）

| 讲义用词 | 正确说法 |
|---|---|
| 切片/Map 是"引用类型" | 切片是**含指针的结构体值**；map 是**指向 hmap 的指针值**。Go 里传参**永远是复制值** |
| "引用传递" | **传指针**，或"共享同一份底层数据" |

用对术语，A13 那个"函数内 append 调用方看不到"的问题就自然讲得通了：复制的是 header，函数内换了指针，调用方的 header 没变。

### 5.7 小结（口述七句）
1. 数组定长且长度属于类型；**默认用切片**。
2. 切片 = ptr + len + cap；传参复制 header，**共享底层数组**。
3. 三索引 `s[low:high:max]` 用容量隔离防污染。
4. `append` 可能重新分配；已知长度用 `make([]T, 0, n)`。
5. Go 1.21+ 用 `slices.Delete/Sort/Contains`，别手写。
6. map 的 key 必须可比较；遍历无序；nil map 可读不可写。
7. 判断存在性用 `v, ok := m[k]`；字段固定用 struct，字段动态才用 map（勘误 A20）。

## 6. 演示脚本

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	// 1) nil 切片 vs 空切片
	var a []int
	b := []int{}
	fmt.Println(len(a), cap(a), a == nil)   // 0 0 true
	fmt.Println(len(b), cap(b), b == nil)   // 0 0 false

	// 2) 观察扩容
	s := []int{4, 5, 6}
	for i := 7; i <= 12; i++ {
		s = append(s, i)
		fmt.Printf("len=%d cap=%d\n", len(s), cap(s))
	}

	// 3) 函数内 append 调用方看不到
	orig := make([]int, 0, 2)
	tryAppend(orig)
	fmt.Println("调用方看到的长度:", len(orig))   // 0

	// 4) 正确写法：返回新切片
	orig = appendAndReturn(orig, 1, 2, 3)
	fmt.Println("返回后:", orig)                  // [1 2 3]

	// 5) map 无序 → 排序遍历
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	fmt.Println("稳定顺序:", keys)                // [a b c]
}

// ❌ 触发扩容后，调用方的 header 没变
func tryAppend(s []int) {
	s = append(s, 1, 2, 3)
}

// ✅ 返回新切片
func appendAndReturn(s []int, vs ...int) []int {
	return append(s, vs...)
}
```

> 对应仓库示例：`s03_array/`、`s04_slice/`、`s06_map/`。

## 7. 课堂练习（含参考答案）

**练习 1** 预测输出：
```go
s := make([]int, 3, 10)
s[0], s[1], s[2] = 1, 2, 3
sub := s[1:3]
sub = append(sub, 99)
fmt.Println(s, sub)
```
> **答**：`s = [1 2 3 99 0 0 0 0 0 0]`，`sub = [2 3 99]`。因为 `sub` 的 cap = 10-1 = 9，append 未扩容，直接写入 `s[3]`。

**练习 2** 写出"删除切片中所有偶数"的两种实现（原地复用 / 新分配），并说明各自的代价。
> **答**：
> ```go
> // 原地（零分配，但破坏原切片）
> func dropEven(s []int) []int {
> 	res := s[:0]
> 	for _, v := range s {
> 		if v%2 != 0 {
> 			res = append(res, v)
> 		}
> 	}
> 	return res
> }
> // 新分配（安全，一次分配）
> func dropEvenSafe(s []int) []int {
> 	res := make([]int, 0, len(s))
> 	for _, v := range s {
> 		if v%2 != 0 {
> 			res = append(res, v)
> 		}
> 	}
> 	return res
> }
> ```
> 代价：原地版会让原切片后半段残留旧值（且共享者被改）；安全版多一次分配。**默认选安全版**，只有在热点路径且确认无共享者时用原地版。

**练习 3** 下面代码有什么问题？怎么改？
```go
var counter map[string]int
func hit(key string) { counter[key]++ }
```
> **答**：`counter` 是 nil map，`counter[key]++` 会 panic（assignment to entry in nil map）。改为 `var counter = make(map[string]int)` 或包级 `func init()` 初始化。

**练习 4** 为什么 `map[[]int]string{}` 编译不过，而 `map[any]string{}` 编译得过却可能运行时 panic？
> **答**：`[]int` 不可比较（切片没有 `==`），编译期就被拒绝（勘误 **A12**）；`any`（`interface{}`）静态类型是"可比较的接口"，编译器放行，但**动态类型**不可比较时在 `hash` 阶段 panic。所以用 `any` 做 key 要自己保证放进去的是可比较类型。

## 8. 勘误与易错预警

| 编号 | 讲义原说法 | 正确说法 |
|---|---|---|
| **A12** | map 的 key/value 可以是任意类型 | **key 必须 comparable**；value 才是任意类型 |
| **A20** | 字段固定时应使用 struct 或 `map[string]interface{}`（性能差） | 字段固定 → struct；字段动态 → map；`map[string]interface{}` 只用于结构确实未知的透传 |
| **C7** | 切片/Map 是引用类型、引用传递 | 含指针的值类型；Go 只有值传递 |
| — | `append` 容量翻倍 | 小切片约翻倍，大切片 ~1.25 倍；**不要依赖具体倍数** |
| — | 学员常写 `s := make([]int, 5)` 再 `append` | 结果是 5 个 0 后面接新元素；要空切片用 `make([]int, 0, 5)` |
| — | 学员用 `for i := 1; i <= len(m); i++` 遍历 map | map 没有下标，key 也不连续；只能 `range` |

## 9. 课后作业（分层）

- **基础**：实现 `func Reverse(s []string) []string`（不修改入参），并写表驱动单测覆盖 nil、空、单元素、奇偶长度。
- **进阶**：实现一个 `GroupBy`：`func GroupBy(users []User, key func(User) string) map[string][]User`，要求所有 value 切片都预分配容量提示。
- **挑战**：实现一个"有序 map"（`OrderedMap`），支持 `Set/Get/Delete/Keys/Len` 与按插入顺序遍历；写单测证明 `Keys()` 顺序稳定。要求线程不安全（并发在 08 篇再处理），但要在文档注释里写清这一点。

## 10. 考核点（可观测）

- [ ] 能画出切片 header 与底层数组的关系图，并标出三索引切片改变了什么。
- [ ] 能预测 `append` 后 `len/cap` 的变化并解释原因。
- [ ] 能解释"函数内 append 调用方看不到"，并给出正确签名。
- [ ] 能写出 map 的稳定遍历代码。
- [ ] 能说出 map key 的 comparable 约束，并举一个运行时 panic 的例子。
- [ ] 基础作业的 `go test ./...` 通过。

## 11. 延伸阅读
- Go 博客：*Go Slices: usage and internals*、*Arrays, slices (and strings): The mechanics of 'append'*
- 标准库：`slices`、`maps`（Go 1.21+）
- 下一课预告：struct 是数据的载体，JSON 是数据的传输格式——以及 `json.Unmarshal` 到 `interface{}` 的经典大坑。
