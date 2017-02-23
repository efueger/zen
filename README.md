# zen is a lightweight go framework for web development

## How to

### Start a server

```go
func main() {
	server := zen.NewServer()

	if err := server.Run(":9999"); err != nil {
		log.Println(err)
	}
}
```

### Add a route

```go
	server := zen.NewServer()
	server.Post("/test", handler)
	server.Get("/test",handler)
	if err := server.Run(":9999"); err != nil {
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
	if err := server.Run(":9999"); err != nil {
	log.Println(err)
	}
```

### Hanndle panic

```go
	server := zen.NewServer()
	server.PanicHandler = handler
	if err := server.Run(":9999"); err != nil {
	log.Println(err)
	}
```

### Hanndle 404

```go
	server := zen.NewServer()
	server.NotFoundHandler = handler
	if err := server.Run(":9999"); err != nil {
	log.Println(err)
	}
```

### Todo

- [ ] Group route
- [ ] Middleware for subpath
- [ ] Grace restart base on go 1.8
- [ ] Optimize performance