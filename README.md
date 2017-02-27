# zen is a lightweight go framework for web development

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![Build Status](https://travis-ci.org/philchia/zen.svg?branch=master)](https://travis-ci.org/philchia/zen)
[![Coverage Status](https://coveralls.io/repos/github/philchia/zen/badge.svg?branch=master)](https://coveralls.io/github/philchia/zen?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.c🗼om/philchia/zen)](https://goreportcard.com/report/github.com/philchia/zen)
[![codebeat badge](https://codebeat.co/badges/fdac6135-0381-45f4-8972-4234f485e6c5)](https://codebeat.co/projects/github-com-philchia-zen-master)
[![GoDoc](https://godoc.org/github.com/philchia/zen?status.svg)](https://godoc.org/github.com/philchia/zen)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

zen is a web framework written by go, you will love it if you preffer high performance and lightweight!!!

⚠️ zen is under heavy development, so the api is not stable so far

## Installation

```bash
go get github.com/philchia/zen
```

## How to

### Start a server

```go
func main() {
	server := zen.NewServer()

	if err := server.Run(":8080"); err != nil {
		log.Println(err)
	}
}
```

### Using GET, POST, PUT, PATCH, DELETE

```go
	server := zen.NewServer()
	server.Get("/test",handler)
	server.Post("/test", handler)
	server.Put("/test",handler)
	server.Patch("/test", handler)
	server.Del("/test",handler)
	if err := server.Run(":8080"); err != nil {
	log.Println(err)
	}
```

### Parameters in path

```go
	server := zen.NewServer()
	server.Get("/user/:uid",func (c *Context) {
		c.JSON(map[string]string{"uid": c.Param(":uid")})
	})
	if err := server.Run(":8080"); err != nil {
	log.Println(err)
	}
```

### Parse and validate input

```go
func handler(c *zen.Context) {
	type Inputs struct {
		Name string `form:"name" json:"name"`
		Age  int    `form:"age" json:"age"`
		Mail string `form:"mail" valid:"[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}" msg:"邮件格式错误" json:"mail"`
	}
	var input Inputs

	if err := c.ParseValidForm(&input); err != nil {
		c.JSON(map[string]string{"err": err.Error()})
		return
	}
	log.Println(input)
	c.JSON(input)
}
```

### Use middleware

```go
	server := zen.NewServer()
	server.Filter(filter)
	if err := server.Run(":8080"); err != nil {
	log.Println(err)
	}
```

### Handle panic

```go
	server := zen.NewServer()
	server.HandlePanic(func(c *zen.Context, err interface{}) {
		c.RawStr(fmt.Sprint(err))
	})
	if err := server.Run(":8080"); err != nil {
	log.Println(err)
	}
```

### Handle 404

```go
	server := zen.NewServer()
	server.HandleNotFound(func(c *zen.Context) {
		c.WriteStatus(http.StatusNotFound)
		c.RawStr("page not found")
	})
	if err := server.Run(":8080"); err != nil {
	log.Println(err)
	}
```

## Todo

- [ ] Detailed Document
- [ ] Handle redirect
- [ ] Unit test
- [ ] More context support
- [ ] Grace restart base on go 1.8
- [ ] fasthttp instead of net/http

## License

zen is published under MIT license