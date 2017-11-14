package mux

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var Param url.Values

type node struct {
	label    byte
	prefix   string
	children []*node
	parent   *node
	params   []string
	method   *requestMethod
}
type requestMethod struct {
	get     http.Handler
	head    http.Handler
	post    http.Handler
	put     http.Handler
	delete  http.Handler
	connect http.Handler
	options http.Handler
	trace   http.Handler
	patch   http.Handler
}

func createRootNode() *node {
	Param = url.Values{}
	return &node{
		label:  '/',
		prefix: "/",
	}
}

// register the handler to the desiered request method
func (m *requestMethod) register(method string, handler http.Handler) {
	switch method {
	case "GET":
		m.get = handler
	case "HEAD":
		m.head = handler
	case "POST":
		m.post = handler
	case "PUT":
		m.put = handler
	case "DELETE":
		m.delete = handler
	case "CONNECT":
		m.connect = handler
	case "OPTIONS":
		m.options = handler
	case "TRACE":
		m.trace = handler
	case "PATCH":
		m.patch = handler
	default:
		panic(fmt.Sprintf("Unknow request method [%s]", method))
	}
}
func (m *requestMethod) getHandler(method string) http.Handler {

	switch method {
	case "GET":
		return m.get
	case "HEAD":
		return m.head
	case "POST":
		return m.post
	case "PUT":
		return m.put
	case "DELETE":
		return m.delete
	case "CONNECT":
		return m.connect
	case "OPTIONS":
		return m.options
	case "TRACE":
		return m.trace
	case "PATCH":
		return m.patch
	default:
		panic(fmt.Sprintf("Unknow request method [%s]", method))
	}
}

// register path into the tree
func (n *node) register(method string, path string, handler http.Handler) {
	if n.prefix == path {
		n.registerHandler(method, handler)
	}
	// Trim the trailing /
	path = strings.TrimRight(path, "/")
	cNode := n
	var (
		cLen    int    // current path length
		pLen    int    // path length
		cpLen   int    // current prefix length
		pPath   string // partial path
		i       int    // match count before static potion
		m       int    // match count
		nPrefix string // next prefix
	)
	cPath := path
	for {
		cLen = len(cPath)
		if cLen == 0 {
			goto End
		}
		cpLen = len(cNode.prefix)
		if cPath[0] != ':' && cPath[0] != '*' {

			// retrieve none static part path
			i = 0
			for ; i < cLen && cPath[i] != ':' && cPath[i] != '*'; i++ {
			}
			pPath = cPath[:i]
			pLen = len(pPath)

		PrefixMatch:
			// try to match the current none static path with current prefix
			m = 0
			for ; m < pLen && m < cpLen && pPath[m] == cNode.prefix[m]; m++ {
			}

			if m == 0 {
				// Case 1: no match
				// create an new child node
				n := &node{label: pPath[0], prefix: pPath}
				n.parent = cNode
				cNode.children = append(cNode.children, n)
				cNode = n
			} else if m == cpLen { // Case 2: matched all current prefix
				if pLen > m {
					// Case 2.1: current none static patch length > matched count
					// Let's split out the none static path into two potions "matched" and "unmatched"
					nPrefix = pPath[m:]

					// Search current Node's child nodes by match the first letter
					n := cNode.findPathByLabel(nPrefix[0])

					if n != nil {
						cNode = n
						// Reset current node's prefix length
						cpLen = len(cNode.prefix)

						// Since we only match partial none static path with current prefix,
						// we need to continue to match rest of potion of none static path
						pPath = pPath[m:]
						pLen = len(pPath)
						goto PrefixMatch
					}

					n = &node{label: nPrefix[0], prefix: nPrefix}
					n.parent = cNode
					cNode.children = append(cNode.children, n)
					cNode = n
				}
			} else if cpLen > m { // Case 3: partially match the current prefix
				nPrefix = cNode.prefix[m:]
				n := &node{label: nPrefix[0], prefix: nPrefix}

				n.children = cNode.children
				n.method = cNode.method
				n.params = cNode.params
				n.parent = cNode
				cNode.prefix = cNode.prefix[:m]
				cNode.method = nil
				cNode.params = nil
				cNode.children = nil
				if pLen > m {
					// Case 3.1 current none static path length > matched count
					nPrefix = pPath[m:]
					nn := &node{label: nPrefix[0], prefix: nPrefix}
					nn.parent = cNode
					// if we append n and nn together, it will only cause one
					// allocation.
					cNode.children = append(cNode.children, n, nn)
					cNode = nn

				} else {
					cNode.children = append(cNode.children, n)
					//cNode = n
				}
			}

		} else {
			//StaticParam:
			if cPath[0] == ':' {
				i = 0
				for ; i < cLen && cPath[i] != '/'; i++ {
				}
				cNode.params = append(cNode.params, cPath[1:i])

			} else if cPath[0] == '*' {
				cNode.params = append(cNode.params, cPath[1:])
			}

		}
		//if (len(cPath) - i) > 1 {

		if len(cPath) > i {

			if cPath[i+1] == ':' {
				cPath = cPath[(i + 1):]
			} else {
				cPath = cPath[i:]
			}
		} else {
			goto End
		}
	}
End:
	if cNode.method == nil {
		cNode.method = new(requestMethod)
	}
	cNode.registerHandler(method, handler)
}

// match request path
// return the handler if found, otherwise, return nil
func (n *node) match(method string, path string) http.Handler {
	if n.prefix == path {
		n.method.getHandler(method)
	}
	// Trim the trailing /
	path = strings.TrimRight(path, "/")
	var (
		cLen int // current path length
		//pLen    int    // path length
		cpLen int // current prefix length
		//		pPath string // partial path
		//i       int    // index for the cutoff the current search path string
		m int // match count
		//nPrefix string // next prefix
	)
	pIndex := 0
	cPath := path[1:]
	cNode := n.findPathByLabel(cPath[0])

	if cNode == nil {
		return nil
	}
	for {
		cLen = len(cPath)
		if cLen == 0 {
			goto End
		}
		cpLen = len(cNode.prefix)
		m = 0
		// match with current prefix
		for ; m < cLen && m < cpLen && cPath[m] == cNode.prefix[m]; m++ {
		}

		//fmt.Println(cNode.prefix)
		if cLen <= m {
			goto End
		}

		// when nothing match and current path length > 0
		if m == 0 {
			goto StaticParam
		}

		// match the current prefix
		if m == cpLen {

			cPath = cPath[m:]
			// Locate the possible child node by matching the label
			n := cNode.findPathByLabel(cPath[0])

			if n != nil {
				// update current node if found
				cNode = n
				continue
			}

			if cPath[0] == '/' {
				cPath = cPath[1:]
			}
		} else {
			return nil
		}

	StaticParam:
		cLen = len(cPath)

		if len(cNode.params) > pIndex {

			m = 0
			for ; m < cLen && cPath[m] != '/'; m++ {
			}
			if cLen > m {
				cPath = cPath[m:]
				Param.Add(cNode.params[pIndex], cPath[:m])
				if len(cNode.params) > 1 {
					pIndex++
				}
			} else {
				Param.Add(cNode.params[pIndex], cPath)
				goto End
			}

		}
	}
End:
	return cNode.method.getHandler(method)
}

// findPathByLabel perform child node search
// Return nil if the current node has no any children or no child node found
// otherwise, return child node if found
func (n *node) findPathByLabel(label byte) *node {
	if n.children == nil {
		return nil
	}
	for _, c := range n.children {
		if c.label == label {
			return c
		}
	}
	return nil
}

func (n *node) registerHandler(method string, handler http.Handler) {
	if n.method == nil {
		n.method = new(requestMethod)
	}
	n.method.register(method, handler)
}
