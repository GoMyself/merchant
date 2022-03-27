package model

import (
	"github.com/beanstalkd/go-beanstalk"
	"github.com/valyala/fasthttp"
	"time"
)

func BeanPut(name string, param map[string]interface{}, delay int) (string, error) {

	m := &fasthttp.Args{}
	for k, v := range param {
		if _, ok := v.(string); ok {
			m.Set(k, v.(string))
		}
	}

	c, err := meta.BeanPool.Get()
	if err != nil {
		return "sys", err
	}

	if conn, ok := c.(*beanstalk.Conn); ok {

		tube := &beanstalk.Tube{Conn: conn, Name: name}
		_, err = tube.Put(m.QueryString(), 1, time.Duration(delay)*time.Second, 10*time.Minute)
		if err != nil {
			_ = meta.BeanPool.Put(c)
			return "sys", err
		}
	}

	//将连接放回连接池中
	return "", meta.BeanPool.Put(c)
}
