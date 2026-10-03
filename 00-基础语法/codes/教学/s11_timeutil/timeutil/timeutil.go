// s11_timeutil/timeutil/timeutil.go —— 对应教学文稿 10 常用标准库速查（一）
package timeutil

import "time"

var (
	cst *time.Location
)

// CSTLayout China Standard Time Layout
const CSTLayout = "2006-01-02 15:04:05"

func init() {
	// 预加载时区，避免 time.LoadLocation 每次调用的开销
	var err error
	if cst, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}
}

// RFC3339ToCSTLayout convert rfc3339 value to china standard time layout
func RFC3339ToCSTLayout(value string) (string, error) {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", err
	}
	return ts.In(cst).Format(CSTLayout), nil
}

// UnixToCSTLayout convert unix timestamp to china standard time layout
func UnixToCSTLayout(ts int64) string {
	return time.Unix(ts, 0).In(cst).Format(CSTLayout)
}
