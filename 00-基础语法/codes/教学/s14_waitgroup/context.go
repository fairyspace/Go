// s14_waitgroup/context.go —— context 取消示例
package main

import "context"

// contextWithCancel 演示 context 的取消机制。
// 真实项目中使用标准库：context.WithCancel(context.Background())
func contextWithCancel() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithCancel(context.Background())
}
