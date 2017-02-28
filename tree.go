package zen

// Handlers is slice of HandlerFunc
type Handlers []HandlerFunc

// Param use to store url key value pair
type Param struct {
	key   string
	value string
}

// Params use to store url parameters
// because slice range is quick than map
type Params []Param

// Get params value by name
func (p Params) Get(name string) string {
	for _, v := range p {
		if v.key == name {
			return v.value
		}
	}
	return ""
}

// countParams return parameter count in path
func countParams(path string) uint8 {
	var ret uint8
	for _, b := range path {
		if b == rune('*') || b == rune(':') {
			ret++
		}
	}
	return ret
}

type nodeType uint8

const (
	root nodeType = iota
	static
	param
	all
)

type node struct {
	ndType    nodeType
	path      string
	children  []*node
	indices   []byte
	prior     int
	handlers  Handlers
	wild      bool
	maxParams uint8
}

type methodNode struct {
	method string
	node   *node
}

func (n *node) increChildsPrior(i int) int {
	n.children[i].prior++
	prio := n.children[i].prior
	// bubble
	newPos := i
	for newPos > 0 && n.children[newPos-1].prior < prio {
		// swap node
		n.children[newPos], n.children[newPos-1] = n.children[newPos-1], n.children[newPos]
		// swap index
		n.indices[newPos], n.indices[newPos-1] = n.indices[newPos-1], n.indices[newPos]
		newPos--
	}
	return newPos
}

func (n *node) addRoute(path string, handlers Handlers) {
	fullpath := path
	// increment n 's priority
	n.prior++
	// get parameters' count in path
	paramsCount := countParams(path)
	// nonempty tree
	if len(n.path) > 0 || len(n.children) > 0 {
	LOOP:
		for {
			// set node's max params count
			n.maxParams = maxUint8(paramsCount, n.maxParams)
			// get longest common prefix index
			i := longestCommonPrefixIndex(path, n.path)

			// cut edge
			if i < len(n.path) {
				child := node{
					path:     n.path[i:],
					wild:     n.wild,
					indices:  n.indices,
					children: n.children,
					handlers: n.handlers,
					prior:    n.prior - 1,
				}

				// update child 's params count
				for i := range child.children {
					child.maxParams = maxUint8(child.maxParams, child.children[i].maxParams)
				}

				n.children = []*node{&child}
				n.indices = []byte{n.path[i]}
				n.path = path[:i]
				n.handlers = nil
				n.wild = false
			}

			// add new node as child of this n
			if i < len(path) {
				path = path[i:]

				if n.wild {
					// n has only 1 child node if n.wild is true
					n = n.children[0]
					n.prior++
					// update child node 's max params count
					n.maxParams = maxUint8(n.maxParams, paramsCount)
					paramsCount--

					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue LOOP
						}
					}
					// conflict path
					panic("path '" + fullpath + "' conflicts with '" + n.path)
				}

				c := path[0]

				if n.ndType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.prior++
					continue LOOP
				}

				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						i = n.increChildsPrior(i)
						n = n.children[i]
						continue LOOP
					}
				}

				if c != ':' && c != '*' {
					n.indices = append(n.indices, c)

					child := &node{
						maxParams: paramsCount,
					}
					n.children = append(n.children, child)
					n.increChildsPrior(len(n.children) - 1)
					n = child
				}
				n.addChild(paramsCount, path, fullpath, handlers)
				return
			} else if i == len(path) {
				if n.handlers != nil {
					panic("duplicate handlers in '" + fullpath + "'")
				}
				n.handlers = handlers
			}
			return
		}
	} else { // empty tree
		n.ndType = root
		n.addChild(paramsCount, path, fullpath, handlers)
	}
}

func (n *node) addChild(numParams uint8, path string, fullPath string, handlers Handlers) {
	var offset int

	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			case ':', '*':
				panic("only one wildcard per path segment is allowed in '" + fullPath + "'")
			default:
				end++
			}
		}

		if len(n.children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		// check if the wildcard has a name
		if end-i < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if c == ':' {
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}
			child := &node{
				ndType:    param,
				maxParams: numParams,
			}
			n.children = []*node{child}
			n.wild = true
			n = child
			n.prior++
			numParams--

			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					prior:     1,
				}
				n.children = []*node{child}
				n = child
			}
		} else {
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wild:      true,
				ndType:    all,
				maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = []byte{path[i]}
			n = child
			n.prior++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				ndType:    all,
				maxParams: 1,
				handlers:  handlers,
				prior:     1,
			}
			n.children = []*node{child}
			return
		}
	}
	n.path = path[offset:]
	n.handlers = handlers
}

func (n *node) get(path string, po Params) (handlers Handlers, p Params) {
	p = po
LOOP:
	//outer loop
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]

				if !n.wild {
					c := path[0]
					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue LOOP
						}
					}

					return
				}

				n = n.children[0]
				switch n.ndType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}
					// save param value
					if cap(p) < int(n.maxParams) {
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].key = n.path[1:]
					p[i].value = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue LOOP
						}
						return
					}

					if handlers = n.handlers; handlers != nil {
						return
					} else if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
					}
					return

				case all:
					// save param value
					if cap(p) < int(n.maxParams) {
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].key = n.path[2:]
					p[i].value = path

					handlers = n.handlers
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == n.path {
			if handlers = n.handlers; handlers != nil {
				return
			}
		}
		return
	}
}
