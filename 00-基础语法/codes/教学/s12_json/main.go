// s12_json/main.go —— 对应教学文稿 04 结构体与 JSON（三）（四）
//
// 运行：go run ./s12_json
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

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

// 情况三：字段不固定，用 squash 嵌入
type Family struct {
	LastName string
}

type Location struct {
	City string
}

type Person struct {
	Family    `json:",squash"`
	Location  `json:",squash"`
	FirstName string
}

// ============ 情况一：结构明确 —— 强类型解析 ============

func strongTyped() {
	fmt.Println("=== 情况一：强类型解析（推荐） ===")
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
	if err := json.Unmarshal([]byte(jsonStr), &mobile); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(mobile.Resultcode)     // 200
	fmt.Println(mobile.Reason)         // Return Successd!
	fmt.Println(mobile.Result.City)    // 杭州
	fmt.Println(mobile.Result.Company) // 中国移动
}

// ============ 情况二：数据类型不确定 —— 弱类型解析（stdlib 手动版） ============
//
// 原始教程使用 github.com/mitchellh/mapstructure 的 WeakDecode()。
// 此处演示等价的手动转换，便于离线运行。
type WeakMobileInfo struct {
	Resultcode string `json:"resultcode"`
}

func weakTyped() {
	fmt.Println("\n=== 情况二：弱类型解析 ===")
	jsonStr := `{"resultcode": 200}`

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Println(err.Error())
		return
	}

	// 手动弱类型转换：float64 -> string
	var mobile WeakMobileInfo
	switch v := result["resultcode"].(type) {
	case float64:
		mobile.Resultcode = fmt.Sprintf("%d", int64(v))
	case string:
		mobile.Resultcode = v
	}

	fmt.Println("定义的是 string，得到：", mobile.Resultcode) // 200
	fmt.Println("实际 JSON 类型：", reflect.TypeOf(result["resultcode"]))
}

// ============ 情况三：字段不固定 ============

func dynamicFields() {
	fmt.Println("\n=== 情况三：字段不固定（squash 摊平） ===")
	jsonStr := `{
		"FirstName": "Mitchell",
		"LastName":  "Hashimoto",
		"City":      "San Francisco"
	}`

	// encoding/json 原生支持匿名内嵌结构体的摊平（无需 squash 标签）
	var result Person
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(result.FirstName, result.LastName, result.City)
	// 说明：mapstructure 需要显式写 `mapstructure:",squash"`，
	// 而 encoding/json 对匿名内嵌字段默认就是摊平的。

	// 未匹配的字段会被忽略
	jsonStr2 := `{"FirstName":"Tom","Unknown":"x"}`
	var r2 Person
	_ = json.Unmarshal([]byte(jsonStr2), &r2)
	fmt.Println("未匹配字段被忽略：", r2.FirstName)
}

// ============ 小坑 1：科学计数法 ============

func pitfallScientific() {
	fmt.Println("\n=== 小坑 1：科学计数法 ===")
	jsonStr := `{"number":1234567}`
	result := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Println(err)
	}
	fmt.Println("现象：", result)                              // map[number:1.234567e+06]
	fmt.Printf("类型：%v\n", reflect.TypeOf(result["number"])) // float64

	// 方案一：强制类型转换
	fmt.Println("方案一 int  ：", int(result["number"].(float64)))
	fmt.Println("方案一 int64：", int64(result["number"].(float64)))
}

// ============ 小坑 1：方案二 定义结构体（推荐） ============

func pitfallSolution2() {
	fmt.Println("\n=== 小坑 1：方案二 定义结构体 ===")
	type Num struct {
		Number int `json:"number"`
	}
	jsonStr := `{"number":1234567}`
	var result Num
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Println(err)
	}
	fmt.Println("结构体接收：", result) // {1234567}

	// 大整数精度丢失演示：雪花 ID
	type ID struct {
		ID int64 `json:"id"`
	}
	snowflake := `{"id":1759482245000000}`
	var i ID
	_ = json.Unmarshal([]byte(snowflake), &i)
	fmt.Println("结构体接收雪花 ID：", i.ID) // 1759482245000000 正确

	var m map[string]interface{}
	_ = json.Unmarshal([]byte(snowflake), &m)
	fmt.Println("map 接收雪花 ID：", m["id"]) // 1.759482245e+15 精度已丢失
}

// ============ 小坑 1：方案三 UseNumber ============

func pitfallSolution3() {
	fmt.Println("\n=== 小坑 1：方案三 UseNumber ===")
	jsonStr := `{"number":1234567}`
	result := make(map[string]interface{})

	d := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	d.UseNumber() // 关键：让数字保持原始字符串形态
	if err := d.Decode(&result); err != nil {
		fmt.Println(err)
	}
	fmt.Println("UseNumber 后：", result) // map[number:1234567]

	// 注意类型变成了 json.Number
	fmt.Printf("类型：%v\n", reflect.TypeOf(result["number"])) // json.Number

	// json.Number 底层是 string，必须显式转换
	numInt, _ := result["number"].(json.Number).Int64()
	fmt.Printf("转 int64：value=%v type=%v\n", numInt, reflect.TypeOf(numInt))

	numStr := result["number"].(json.Number).String()
	fmt.Printf("转 string：value=%v type=%v\n", numStr, reflect.TypeOf(numStr))

	// 陷阱：json.Number 是字符串，直接参与算术会编译报错
	// numInt + 1  // 编译错误：invalid operation
	// 正确写法：
	fmt.Println("参与算术前先转换：", numInt+1)
}

func main() {
	strongTyped()
	weakTyped()
	dynamicFields()
	pitfallScientific()
	pitfallSolution2()
	pitfallSolution3()
}
