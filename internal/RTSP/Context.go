package RTSP

import "net/url"

type Context struct {
	url    *url.URL
	method string
	req    Request
	resp   Response

	Keys map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		Keys: make(map[string]interface{}),
	}
}
