package zen

// parseParams parse url parameters
func (r *route) parseParams(c *Context, parts []string) {

	for i, k := range r.params {
		c.params[k] = parts[i]
	}
}
