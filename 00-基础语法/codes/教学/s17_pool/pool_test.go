// s17_pool/pool_test.go —— 用 testing.B 对比「每次 new」与「sync.Pool 复用」
//
// 运行：go test -bench . -benchmem ./s17_pool
//
// 为什么不用 time.Now() 手写微基准：
//  1. 短代码可能被编译器整体优化掉（对象根本不会被真正分配）；
//  2. 没有多轮取样，单次结果噪声大；
//  3. 拿不到 allocs/op —— 而分配次数才是判断 GC 压力的首要指标。
package main

import "testing"

// sink 防止编译器把「未使用的对象」优化掉
var sink *student

// BenchmarkNewStudent 每次都新建对象
func BenchmarkNewStudent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := &student{Name: "Tom", Age: 30}
		s.Tags = append(s.Tags, "读书", "爬山")
		sink = s
	}
}

// BenchmarkPooledStudent 用 sync.Pool 复用（Get 后必须 Release）
func BenchmarkPooledStudent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := New("Tom", 30)
		s.Tags = append(s.Tags, "读书", "爬山")
		sink = s
		Release(s) // 内部会重置 Name/Age/Tags 后归还
	}
}

// TestReleaseResets 验证 Release 确实重置了字段。
// 池化对象最大的风险是「漏重置字段」导致脏数据串到下一个调用者，这条测试就是那道保险。
//
// 注意：这里在 Release 之后仍然读取 s 的字段，只是为了断言重置结果，
// 且测试是单 goroutine 的；在并发程序中，Put 之后绝不能再碰这个对象。
func TestReleaseResets(t *testing.T) {
	s := New("Tom", 30)
	s.Tags = append(s.Tags, "读书", "爬山")
	Release(s)

	if s.Name != "" || s.Age != 0 || s.Tags != nil {
		t.Errorf("Release 未重置干净: name=%q age=%d tags=%v", s.Name, s.Age, s.Tags)
	}
}
