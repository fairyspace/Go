// s06_map/main.go —— 对应教学文稿 03 数组、切片与 Map（三）映射 Map
//
// 运行：go run ./s06_map
package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

func main() {
	// ==================== 声明 Map：六种等价写法 ====================
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

	for i, m := range []map[int]string{p1, p2, p3, p4, p5, p6} {
		fmt.Printf("写法%d：len=%d 内容=%v\n", i+1, len(m), m)
	}

	// ==================== nil map 的陷阱 ====================
	fmt.Println("\n--- nil map ---")
	var nilMap map[string]int
	fmt.Println("读取安全：", nilMap["a"])   // 0
	delete(nilMap, "a")                 // 删除安全
	fmt.Println("删除后仍可读：", nilMap["a"]) // 0
	// nilMap["a"] = 1                          // panic: assignment to entry in nil map

	// ==================== 嵌套 Map 与 JSON ====================
	res := make(map[string]interface{})
	res["code"] = 200
	res["msg"] = "success"
	res["data"] = map[string]interface{}{
		"username": "Tom",
		"age":      "30",
		"hobby":    []string{"读书", "爬山"},
	}
	fmt.Println("\nmap data :", res)

	// 序列化
	jsons, errs := json.Marshal(res)
	if errs != nil {
		fmt.Println("json marshal error:", errs)
	}
	fmt.Println("--- map to json ---")
	fmt.Println("json data :", string(jsons))

	// 反序列化
	res2 := make(map[string]interface{})
	if err := json.Unmarshal(jsons, &res2); err != nil {
		fmt.Println("json unmarshal error:", err)
	}
	fmt.Println("--- json to map ---")
	fmt.Println("map data :", res2)
	// 注意：res2 中的数字全部变成了 float64（见第 04 篇）

	// ==================== 编辑和删除 ====================
	fmt.Println("\n--- 编辑删除 ---")
	person := map[int]string{
		1: "Tom",
		2: "Aaron",
		3: "John",
	}
	fmt.Println("data :", person)

	person[2] = "Jack"  // 修改
	person[4] = "Kevin" // 新增
	fmt.Println("data :", person)

	delete(person, 2)  // 删除
	delete(person, 99) // 删除不存在的 key 不会报错
	fmt.Println("data :", person)

	// ==================== 判断 key 是否存在 ====================
	fmt.Println("\n--- 判断存在性 ---")
	val, ok := person[1]
	if ok {
		fmt.Println("存在，值为", val)
	}
	// 与之配套，val == "" 不能用来判断 key 是否存在
	fmt.Println("person[1] 是否为空串：", person[1] == "")

	// ==================== 遍历 ====================
	fmt.Println("\n--- 遍历（顺序随机）---")
	for k, v := range person {
		fmt.Printf("person[%d]: %s\n", k, v)
	}

	// 需要稳定顺序时，先排序 key
	keys := make([]int, 0, len(person))
	for k := range person {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("有序 person[%d]: %s\n", k, person[k])
	}
}
