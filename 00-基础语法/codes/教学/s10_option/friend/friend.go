// s10_option/friend/friend.go —— 对应教学文稿 07 接口与 Options 模式（二）
package friend

import (
	"fmt"
)

// Find 使用 Options 模式查找
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

// ============ 对比：结构体参数写法 ============
//
// 配置项少于 3 个、或必须全部指定时，结构体参数是更好的选择：
// 它保留了编译期字段名检查，代码更短也更易读。
type FindOptions struct {
	Sex    int
	Age    int
	Height int
	Weight int
	Hobby  string
}

func FindWithStruct(where string, opts FindOptions) (string, error) {
	res := fmt.Sprintf("从 %s 找朋友\n", where)
	if opts.Sex == 1 {
		res += "性别：女性\n"
	}
	if opts.Age != 0 {
		res += fmt.Sprintf("年龄：%d岁\n", opts.Age)
	}
	if opts.Hobby != "" {
		res += fmt.Sprintf("爱好：%s\n", opts.Hobby)
	}
	return res, nil
}
