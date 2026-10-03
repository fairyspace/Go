// s11_timeutil/main.go —— 对应教学文稿 10 常用标准库速查（一）（二）
//
// 运行：go run ./s11_timeutil
package main

import (
	"demo/s11_timeutil/timeutil"
	"fmt"
	"time"
)

func main() {
	// ==================== RFC3339 转 CST ====================
	RFC3339Str := "2020-11-08T08:18:46+08:00"
	cst, err := timeutil.RFC3339ToCSTLayout(RFC3339Str)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("--- RFC3339 转 CST ---")
	fmt.Println(RFC3339Str, "->", cst)

	// 解析失败
	_, err = timeutil.RFC3339ToCSTLayout("invalid")
	fmt.Println("解析失败：", err)

	// ==================== time 常用操作 ====================
	fmt.Println("\n--- time 常用 ---")
	now := time.Now()
	fmt.Println("now          =", now)
	fmt.Println("Year/Month   =", now.Year(), now.Month())
	fmt.Println("Day/Weekday  =", now.Day(), now.Weekday())
	fmt.Println("Unix         =", now.Unix())
	fmt.Println("UnixMilli    =", now.UnixMilli())
	fmt.Println("UnixNano     =", now.UnixNano())

	// 时间戳 → 时间对象
	t := time.Unix(1759482245, 0)
	fmt.Println("Unix 转时间   =", t.Format("2006-01-02 15:04:05"))
	fmt.Println("UnixMilli 转  =", timeutil.UnixToCSTLayout(1759482245))

	// 字符串 → 时间对象
	t2, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-11-08 08:18:46", time.Local)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("解析字符串    =", t2)

	// 时间运算
	fmt.Println("Add 24h       =", now.Add(24*time.Hour))
	fmt.Println("AddDate 1月   =", now.AddDate(0, 1, 0))
	fmt.Println("Add -2h       =", now.Add(-2*time.Hour))
	fmt.Println("Sub(now - t2) =", now.Sub(t2))

	// 时间比较
	fmt.Println("After  t2     =", now.After(t2))
	fmt.Println("Before t2     =", now.Before(t2))
	fmt.Println("Equal   t2    =", now.Equal(t2))

	// 格式化
	fmt.Println("\n--- 格式化 ---")
	fmt.Println(now.Format("2006-01-02 15:04:05"))
	fmt.Println(now.Format("2006-01-02"))
	fmt.Println(now.Format("15:04:05"))
	fmt.Println(now.Format(time.RFC3339))
	fmt.Println(now.Format("2006年01月02日 15时04分05秒"))

	// 定时器
	fmt.Println("\n--- 定时器 ---")
	timer := time.NewTimer(100 * time.Millisecond)
	<-timer.C
	fmt.Println("Timer 已触发")

	ticker := time.NewTicker(50 * time.Millisecond)
	count := 0
	for range ticker.C {
		count++
		if count == 3 {
			ticker.Stop()
			fmt.Println("Ticker 触发 3 次后停止")
			break
		}
	}

	// 超时
	select {
	case <-time.After(50 * time.Millisecond):
		fmt.Println("select 超时分支")
	default:
		fmt.Println("select default 分支")
	}
}
