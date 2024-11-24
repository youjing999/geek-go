package geecache

// 缓存值的抽象与封装

// ByteView b将存储真实的缓存值，选择byte类型为了支持任意的数据类型的存储
type ByteView struct {
	b []byte
}

// Len 要求被缓存对象必须实现 Value 接口
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 将[]byte 转化为字符串
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
