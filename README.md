# zen is a lightweight go framework for web development

## How to

### Start a server

```go
func main() {
	server := zen.NewServer()

	server.Post("/test", handler)
	if err := server.Run(":9999"); err != nil {
		log.Println(err)
	}
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