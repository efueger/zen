package zen

import (
	"log"
	"strings"
)

func (r *route) parseParams(c *Context) {
	pattern := c.Req.URL.Path
	pattern = strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/")
	parts := strings.Split(pattern, "/")
	log.Println("parts", parts)
	log.Println(r.params)
	for i, k := range r.params {
		c.params[k] = parts[i]
	}
}
