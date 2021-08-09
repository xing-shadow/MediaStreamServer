package RTMP

type Context struct {
	Keys map[string]interface{}
}

func NewContext() *Context {
	c := &Context{
		Keys: make(map[string]interface{}),
	}
	return c
}
