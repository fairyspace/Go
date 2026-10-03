// s09_interface/study/study.go —— 对应教学文稿 07 接口与 Options 模式（一）
package study

import "errors"

// 编译期断言：要求 *study 必须实现 Study
// 若 Study 接口被更改或未全部实现，这里会编译报错
var _ Study = (*study)(nil)

// Study 接口
type Study interface {
	Listen(msg string) string
	Speak(msg string) string
	Read(msg string) string
	Write(msg string) string
}

// 私有结构体：不想在其他地方被使用
type study struct {
	Name string
}

func (s *study) Listen(msg string) string { return s.Name + " 听 " + msg }
func (s *study) Speak(msg string) string  { return s.Name + " 说 " + msg }
func (s *study) Read(msg string) string   { return s.Name + " 读 " + msg }
func (s *study) Write(msg string) string  { return s.Name + " 写 " + msg }

// New 构造函数：返回接口而非具体类型
func New(name string) (Study, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	return &study{
		Name: name,
	}, nil
}

// ============ 接口组合 ============

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

// 组合成新的接口
type ReadWriter interface {
	Reader
	Writer
}

// 一个类型可以同时实现多个接口
type File struct {
	Data []byte
}

func (f *File) Read(p []byte) (int, error) {
	n := copy(p, f.Data)
	return n, nil
}

func (f *File) Write(p []byte) (int, error) {
	f.Data = append(f.Data, p...)
	return len(p), nil
}

// 编译期断言：*File 实现了 ReadWriter
var _ ReadWriter = (*File)(nil)
