package mux

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestRegisterOne(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", nil)
	assert.Equal(t, "/", tree.prefix)
	assert.Equal(t, byte('v'), tree.children[0].label)
	assert.Equal(t, "v1/test", tree.children[0].prefix)
}

func TestRegisterSplitNodeInToTwo(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", nil)
	assert.Equal(t, "v1/test", tree.children[0].prefix)
	tree.register("GET", "/v1/tett", nil)
	assert.Equal(t, "v1/te", tree.children[0].prefix)
	assert.Equal(t, "st", tree.children[0].children[0].prefix)
	assert.Equal(t, "tt", tree.children[0].children[1].prefix)
}

func TestRegisterParam(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test/:id", nil)
	assert.Equal(t, "v1/test/", tree.children[0].prefix)
	assert.Equal(t, "id", tree.children[0].params[0])
}

func TestRegisterConsecutiveParams(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test/:id/:name/:phone", nil)
	assert.Equal(t, "v1/test/", tree.children[0].prefix)
	assert.Equal(t, 3, len(tree.children[0].params))
	assert.Equal(t, "id,name,phone", strings.Join(tree.children[0].params, ","))
}

func TestRegisterMultipleNotConsecutiveParams(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test/:id/name/:phone", nil)
	assert.Equal(t, "v1/test/", tree.children[0].prefix)
	assert.Equal(t, 1, len(tree.children[0].params))
	assert.Equal(t, "id", tree.children[0].params[0])
	assert.Equal(t, "/name/", tree.children[0].children[0].prefix)
	assert.Equal(t, "phone", tree.children[0].children[0].params[0])
}

var fakeHandlerValue string

func fakeHandler(val string) http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		fakeHandlerValue = val
	})
}

func TestMatch(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", fakeHandler("/v1/test"))
	h := tree.match("GET", "/v1/test")
	h.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/test", fakeHandlerValue)
}

func TtetMatchSingleParam(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", fakeHandler("/v1/test"))
	tree.register("GET", "/v1/:id", fakeHandler("/v1/:id"))
	h := tree.match("GET", "/v1/1")
	h.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/:id", fakeHandlerValue)
}

func TtetMatchConsecutiveParams(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", fakeHandler("/v1/test"))
	tree.register("GET", "/v1/:id/:name/:account_id", fakeHandler("/v1/:id/:name/:account_id"))
	h := tree.match("GET", "/v1/1/2/3")
	h.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/:id/:name:/:account_id", fakeHandlerValue)

	h1 := tree.match("GET", "/v1/test")
	h1.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/test", fakeHandlerValue)
}

func TestMatchMultipleNotConsecutiveParams(t *testing.T) {
	tree := createRootNode()
	tree.register("GET", "/v1/test", fakeHandler("/v1/test"))
	tree.register("GET", "/v1/test/find/:id", fakeHandler("/v1/test/find/:id"))

	tree.register("GET", "/v1/test/:account_id/loan/:loan_id", fakeHandler("/v1/test/:account_id/loan/:loan_id"))
	h1 := tree.match("GET", "/v1/test")
	h2 := tree.match("GET", "/v1/test/find/1")
	h3 := tree.match("GET", "/v1/test/1/loan/2")
	h1.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/test", fakeHandlerValue)
	h2.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/test/find/:id", fakeHandlerValue)
	h3.ServeHTTP(nil, nil)
	assert.Equal(t, "/v1/test/:account_id/loan/:loan_id", fakeHandlerValue)
}
