// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE

package gin

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Used as a workaround since we can't compare functions or their addresses
var fakeHandlerValue string

func fakeHandler(val string) HandlersChain {
	return HandlersChain{func(c *Context) {
		fakeHandlerValue = val
	}}
}

type testRequests []struct {
	path       string
	nilHandler bool
	route      string
	ps         Params
}

func getParams() *Params {
	ps := make(Params, 0, 20)
	return &ps
}

func getSkippedNodes() *[]skippedNode {
	ps := make([]skippedNode, 0, 20)
	return &ps
}

func checkRequests(t *testing.T, tree *node, requests testRequests, unescapes ...bool) {
	unescape := false
	if len(unescapes) >= 1 {
		unescape = unescapes[0]
	}

	for _, request := range requests {
		value := tree.getValue(request.path, getParams(), getSkippedNodes(), unescape)

		if value.handlers == nil {
			if !request.nilHandler {
				t.Errorf("handle mismatch for route '%s': Expected non-nil handle", request.path)
			}
		} else if request.nilHandler {
			t.Errorf("handle mismatch for route '%s': Expected nil handle", request.path)
		} else {
			value.handlers[0](nil)
			if fakeHandlerValue != request.route {
				t.Errorf("handle mismatch for route '%s': Wrong handle (%s != %s)", request.path, fakeHandlerValue, request.route)
			}
		}

		if value.params != nil {
			if !reflect.DeepEqual(*value.params, request.ps) {
				t.Errorf("Params mismatch for route '%s'", request.path)
			}
		}

	}
}

func checkPriorities(t *testing.T, n *node) uint32 {
	var prio uint32
	for i := range n.children {
		prio += checkPriorities(t, n.children[i])
	}

	if n.handlers != nil {
		prio++
	}

	if n.priority != prio {
		t.Errorf(
			"priority mismatch for node '%s': is %d, should be %d",
			n.path, n.priority, prio,
		)
	}

	return prio
}

func TestCountParams(t *testing.T) {
	if countParams("/path/:param1/static/*catch-all") != 2 {
		t.Fail()
	}
	if countParams(strings.Repeat("/:param", 256)) != 256 {
		t.Fail()
	}
}

func TestTreeAddAndGet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		literal_4217,
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		literal_2035,
		literal_0261,
		"/α",
		"/β",
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler(route))
	}

	checkRequests(t, tree, testRequests{
		{"/a", false, "/a", nil},
		{"/", true, "", nil},
		{"/hi", false, "/hi", nil},
		{literal_4217, false, literal_4217, nil},
		{"/co", false, "/co", nil},
		{"/con", true, "", nil},  // key mismatch
		{"/cona", true, "", nil}, // key mismatch
		{"/no", true, "", nil},   // no matching child
		{"/ab", false, "/ab", nil},
		{"/α", false, "/α", nil},
		{"/β", false, "/β", nil},
	})

	checkPriorities(t, tree)
}

func TestTreeWildcard(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		literal_0946,
		literal_8372,
		literal_5389,
		literal_6049,
		literal_4827,
		literal_2791,
		literal_1489,
		literal_8590,
		literal_8605,
		literal_0652,
		literal_0836,
		"/user_:name/about",
		literal_6470,
		"/doc/",
		literal_2035,
		literal_0261,
		"/info/:user/public",
		literal_8420,
		"/info/:user/project/golang",
		"/aa/*xx",
		"/ab/*xx",
		"/:cc",
		"/c1/:dd/e",
		"/c1/:dd/e1",
		literal_2845,
		literal_5382,
		"/:cc/:dd/:ee/ff",
		"/:cc/:dd/:ee/:ff/gg",
		literal_0384,
		literal_2645,
		literal_0297,
		literal_4176,
		literal_4691,
		literal_5986,
		literal_0974,
		literal_8120,
		literal_9526,
		literal_3065,
		literal_3026,
		literal_6074,
		literal_4320,
		literal_9802,
		literal_0125,
		literal_6970,
		literal_6843,
		literal_9516,
		"/get/abc/123abd/:param",
		"/get/abc/123abddd/:param",
		"/get/abc/123/:param",
		"/get/abc/123abg/:param",
		"/get/abc/123abf/:param",
		"/get/abc/123abfff/:param",
		literal_3654,
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler(route))
	}

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/cmd/test", true, literal_0946, Params{Param{"tool", "test"}}},
		{"/cmd/test/", false, literal_0946, Params{Param{"tool", "test"}}},
		{"/cmd/test/3", false, literal_8372, Params{Param{Key: "tool", Value: "test"}, Param{Key: "sub", Value: "3"}}},
		{"/cmd/who", true, literal_0946, Params{Param{"tool", "who"}}},
		{"/cmd/who/", false, literal_0946, Params{Param{"tool", "who"}}},
		{literal_5389, false, literal_5389, nil},
		{"/cmd/whoami/", true, literal_5389, nil},
		{"/cmd/whoami/r", false, literal_8372, Params{Param{Key: "tool", Value: "whoami"}, Param{Key: "sub", Value: "r"}}},
		{"/cmd/whoami/r/", true, literal_8372, Params{Param{Key: "tool", Value: "whoami"}, Param{Key: "sub", Value: "r"}}},
		{literal_6049, false, literal_6049, nil},
		{literal_4827, false, literal_4827, nil},
		{"/src/", false, literal_2791, Params{Param{Key: "filepath", Value: "/"}}},
		{literal_3619, false, literal_2791, Params{Param{Key: "filepath", Value: "/some/file.png"}}},
		{literal_1489, false, literal_1489, nil},
		{literal_7309, false, literal_8590, Params{Param{Key: "query", Value: "someth!ng+in+ünìcodé"}}},
		{"/search/someth!ng+in+ünìcodé/", true, "", Params{Param{Key: "query", Value: "someth!ng+in+ünìcodé"}}},
		{"/search/gin", false, literal_8590, Params{Param{"query", "gin"}}},
		{literal_8605, false, literal_8605, nil},
		{literal_0652, false, literal_0652, nil},
		{"/user_gopher", false, literal_0836, Params{Param{Key: "name", Value: "gopher"}}},
		{"/user_gopher/about", false, "/user_:name/about", Params{Param{Key: "name", Value: "gopher"}}},
		{"/files/js/inc/framework.js", false, literal_6470, Params{Param{Key: "dir", Value: "js"}, Param{Key: "filepath", Value: "/inc/framework.js"}}},
		{"/info/gordon/public", false, "/info/:user/public", Params{Param{Key: "user", Value: "gordon"}}},
		{"/info/gordon/project/go", false, literal_8420, Params{Param{Key: "user", Value: "gordon"}, Param{Key: "project", Value: "go"}}},
		{"/info/gordon/project/golang", false, "/info/:user/project/golang", Params{Param{Key: "user", Value: "gordon"}}},
		{"/aa/aa", false, "/aa/*xx", Params{Param{Key: "xx", Value: "/aa"}}},
		{"/ab/ab", false, "/ab/*xx", Params{Param{Key: "xx", Value: "/ab"}}},
		{"/a", false, "/:cc", Params{Param{Key: "cc", Value: "a"}}},
		// * Error with argument being intercepted
		// new PR handle (/all /all/cc /a/cc)
		// fix PR: https://github.com/jialequ/mpgw/pull/2796
		{"/all", false, "/:cc", Params{Param{Key: "cc", Value: "all"}}},
		{"/d", false, "/:cc", Params{Param{Key: "cc", Value: "d"}}},
		{"/ad", false, "/:cc", Params{Param{Key: "cc", Value: "ad"}}},
		{"/dd", false, "/:cc", Params{Param{Key: "cc", Value: "dd"}}},
		{"/dddaa", false, "/:cc", Params{Param{Key: "cc", Value: "dddaa"}}},
		{"/aa", false, "/:cc", Params{Param{Key: "cc", Value: "aa"}}},
		{"/aaa", false, "/:cc", Params{Param{Key: "cc", Value: "aaa"}}},
		{"/aaa/cc", false, literal_2845, Params{Param{Key: "cc", Value: "aaa"}}},
		{"/ab", false, "/:cc", Params{Param{Key: "cc", Value: "ab"}}},
		{"/abb", false, "/:cc", Params{Param{Key: "cc", Value: "abb"}}},
		{"/abb/cc", false, literal_2845, Params{Param{Key: "cc", Value: "abb"}}},
		{"/allxxxx", false, "/:cc", Params{Param{Key: "cc", Value: "allxxxx"}}},
		{"/alldd", false, "/:cc", Params{Param{Key: "cc", Value: "alldd"}}},
		{"/all/cc", false, literal_2845, Params{Param{Key: "cc", Value: "all"}}},
		{"/a/cc", false, literal_2845, Params{Param{Key: "cc", Value: "a"}}},
		{"/c1/d/e", false, "/c1/:dd/e", Params{Param{Key: "dd", Value: "d"}}},
		{"/c1/d/e1", false, "/c1/:dd/e1", Params{Param{Key: "dd", Value: "d"}}},
		{"/c1/d/ee", false, literal_5382, Params{Param{Key: "cc", Value: "c1"}, Param{Key: "dd", Value: "d"}}},
		{"/cc/cc", false, literal_2845, Params{Param{Key: "cc", Value: "cc"}}},
		{"/ccc/cc", false, literal_2845, Params{Param{Key: "cc", Value: "ccc"}}},
		{"/deedwjfs/cc", false, literal_2845, Params{Param{Key: "cc", Value: "deedwjfs"}}},
		{"/acllcc/cc", false, literal_2845, Params{Param{Key: "cc", Value: "acllcc"}}},
		{literal_2645, false, literal_2645, nil},
		{"/get/te/abc/", false, literal_0297, Params{Param{Key: "param", Value: "te"}}},
		{"/get/testaa/abc/", false, literal_0297, Params{Param{Key: "param", Value: "testaa"}}},
		{"/get/xx/abc/", false, literal_0297, Params{Param{Key: "param", Value: "xx"}}},
		{"/get/tt/abc/", false, literal_0297, Params{Param{Key: "param", Value: "tt"}}},
		{"/get/a/abc/", false, literal_0297, Params{Param{Key: "param", Value: "a"}}},
		{"/get/t/abc/", false, literal_0297, Params{Param{Key: "param", Value: "t"}}},
		{"/get/aa/abc/", false, literal_0297, Params{Param{Key: "param", Value: "aa"}}},
		{"/get/abas/abc/", false, literal_0297, Params{Param{Key: "param", Value: "abas"}}},
		{literal_4691, false, literal_4691, nil},
		{"/something/abcdad/thirdthing", false, literal_4176, Params{Param{Key: "paramname", Value: "abcdad"}}},
		{"/something/secondthingaaaa/thirdthing", false, literal_4176, Params{Param{Key: "paramname", Value: "secondthingaaaa"}}},
		{"/something/se/thirdthing", false, literal_4176, Params{Param{Key: "paramname", Value: "se"}}},
		{"/something/s/thirdthing", false, literal_4176, Params{Param{Key: "paramname", Value: "s"}}},
		{"/c/d/ee", false, literal_5382, Params{Param{Key: "cc", Value: "c"}, Param{Key: "dd", Value: "d"}}},
		{"/c/d/e/ff", false, "/:cc/:dd/:ee/ff", Params{Param{Key: "cc", Value: "c"}, Param{Key: "dd", Value: "d"}, Param{Key: "ee", Value: "e"}}},
		{"/c/d/e/f/gg", false, "/:cc/:dd/:ee/:ff/gg", Params{Param{Key: "cc", Value: "c"}, Param{Key: "dd", Value: "d"}, Param{Key: "ee", Value: "e"}, Param{Key: "ff", Value: "f"}}},
		{"/c/d/e/f/g/hh", false, literal_0384, Params{Param{Key: "cc", Value: "c"}, Param{Key: "dd", Value: "d"}, Param{Key: "ee", Value: "e"}, Param{Key: "ff", Value: "f"}, Param{Key: "gg", Value: "g"}}},
		{"/cc/dd/ee/ff/gg/hh", false, literal_0384, Params{Param{Key: "cc", Value: "cc"}, Param{Key: "dd", Value: "dd"}, Param{Key: "ee", Value: "ee"}, Param{Key: "ff", Value: "ff"}, Param{Key: "gg", Value: "gg"}}},
		{literal_5986, false, literal_5986, nil},
		{"/get/a", false, literal_0974, Params{Param{Key: "param", Value: "a"}}},
		{"/get/abz", false, literal_0974, Params{Param{Key: "param", Value: "abz"}}},
		{"/get/12a", false, literal_0974, Params{Param{Key: "param", Value: "12a"}}},
		{"/get/abcd", false, literal_0974, Params{Param{Key: "param", Value: "abcd"}}},
		{literal_8120, false, literal_8120, nil},
		{"/get/abc/12", false, literal_9526, Params{Param{Key: "param", Value: "12"}}},
		{"/get/abc/123ab", false, literal_9526, Params{Param{Key: "param", Value: "123ab"}}},
		{"/get/abc/xyz", false, literal_9526, Params{Param{Key: "param", Value: "xyz"}}},
		{"/get/abc/123abcddxx", false, literal_9526, Params{Param{Key: "param", Value: "123abcddxx"}}},
		{literal_3065, false, literal_3065, nil},
		{"/get/abc/123abc/x", false, literal_3026, Params{Param{Key: "param", Value: "x"}}},
		{"/get/abc/123abc/xxx", false, literal_3026, Params{Param{Key: "param", Value: "xxx"}}},
		{"/get/abc/123abc/abc", false, literal_3026, Params{Param{Key: "param", Value: "abc"}}},
		{"/get/abc/123abc/xxx8xxas", false, literal_3026, Params{Param{Key: "param", Value: "xxx8xxas"}}},
		{literal_6074, false, literal_6074, nil},
		{"/get/abc/123abc/xxx8/1", false, literal_4320, Params{Param{Key: "param", Value: "1"}}},
		{"/get/abc/123abc/xxx8/123", false, literal_4320, Params{Param{Key: "param", Value: "123"}}},
		{"/get/abc/123abc/xxx8/78k", false, literal_4320, Params{Param{Key: "param", Value: "78k"}}},
		{"/get/abc/123abc/xxx8/1234xxxd", false, literal_4320, Params{Param{Key: "param", Value: "1234xxxd"}}},
		{literal_9802, false, literal_9802, nil},
		{"/get/abc/123abc/xxx8/1234/f", false, literal_0125, Params{Param{Key: "param", Value: "f"}}},
		{"/get/abc/123abc/xxx8/1234/ffa", false, literal_0125, Params{Param{Key: "param", Value: "ffa"}}},
		{"/get/abc/123abc/xxx8/1234/kka", false, literal_0125, Params{Param{Key: "param", Value: "kka"}}},
		{"/get/abc/123abc/xxx8/1234/ffas321", false, literal_0125, Params{Param{Key: "param", Value: "ffas321"}}},
		{literal_6970, false, literal_6970, nil},
		{"/get/abc/123abc/xxx8/1234/kkdd/1", false, literal_6843, Params{Param{Key: "param", Value: "1"}}},
		{"/get/abc/123abc/xxx8/1234/kkdd/12", false, literal_6843, Params{Param{Key: "param", Value: "12"}}},
		{"/get/abc/123abc/xxx8/1234/kkdd/12b", false, literal_6843, Params{Param{Key: "param", Value: "12b"}}},
		{"/get/abc/123abc/xxx8/1234/kkdd/34", false, literal_6843, Params{Param{Key: "param", Value: "34"}}},
		{"/get/abc/123abc/xxx8/1234/kkdd/12c2e3", false, literal_6843, Params{Param{Key: "param", Value: "12c2e3"}}},
		{"/get/abc/12/test", false, literal_9516, Params{Param{Key: "param", Value: "12"}}},
		{"/get/abc/123abdd/test", false, literal_9516, Params{Param{Key: "param", Value: "123abdd"}}},
		{"/get/abc/123abdddf/test", false, literal_9516, Params{Param{Key: "param", Value: "123abdddf"}}},
		{"/get/abc/123ab/test", false, literal_9516, Params{Param{Key: "param", Value: "123ab"}}},
		{"/get/abc/123abgg/test", false, literal_9516, Params{Param{Key: "param", Value: "123abgg"}}},
		{"/get/abc/123abff/test", false, literal_9516, Params{Param{Key: "param", Value: "123abff"}}},
		{"/get/abc/123abffff/test", false, literal_9516, Params{Param{Key: "param", Value: "123abffff"}}},
		{"/get/abc/123abd/test", false, "/get/abc/123abd/:param", Params{Param{Key: "param", Value: "test"}}},
		{"/get/abc/123abddd/test", false, "/get/abc/123abddd/:param", Params{Param{Key: "param", Value: "test"}}},
		{"/get/abc/123/test22", false, "/get/abc/123/:param", Params{Param{Key: "param", Value: "test22"}}},
		{"/get/abc/123abg/test", false, "/get/abc/123abg/:param", Params{Param{Key: "param", Value: "test"}}},
		{"/get/abc/123abf/testss", false, "/get/abc/123abf/:param", Params{Param{Key: "param", Value: "testss"}}},
		{"/get/abc/123abfff/te", false, "/get/abc/123abfff/:param", Params{Param{Key: "param", Value: "te"}}},
		{literal_3654, false, literal_3654, nil},
	})

	checkPriorities(t, tree)
}

func TestUnescapeParameters(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		literal_8372,
		literal_0946,
		literal_2791,
		literal_8590,
		literal_6470,
		literal_8420,
		literal_4578,
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler(route))
	}

	unescape := true
	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/cmd/test/", false, literal_0946, Params{Param{Key: "tool", Value: "test"}}},
		{"/cmd/test", true, "", Params{Param{Key: "tool", Value: "test"}}},
		{literal_3619, false, literal_2791, Params{Param{Key: "filepath", Value: "/some/file.png"}}},
		{"/src/some/file+test.png", false, literal_2791, Params{Param{Key: "filepath", Value: "/some/file test.png"}}},
		{"/src/some/file++++%%%%test.png", false, literal_2791, Params{Param{Key: "filepath", Value: "/some/file++++%%%%test.png"}}},
		{"/src/some/file%2Ftest.png", false, literal_2791, Params{Param{Key: "filepath", Value: "/some/file/test.png"}}},
		{literal_7309, false, literal_8590, Params{Param{Key: "query", Value: "someth!ng in ünìcodé"}}},
		{"/info/gordon/project/go", false, literal_8420, Params{Param{Key: "user", Value: "gordon"}, Param{Key: "project", Value: "go"}}},
		{"/info/slash%2Fgordon", false, literal_4578, Params{Param{Key: "user", Value: "slash/gordon"}}},
		{"/info/slash%2Fgordon/project/Project%20%231", false, literal_8420, Params{Param{Key: "user", Value: "slash/gordon"}, Param{Key: "project", Value: "Project #1"}}},
		{"/info/slash%%%%", false, literal_4578, Params{Param{Key: "user", Value: "slash%%%%"}}},
		{"/info/slash%%%%2Fgordon/project/Project%%%%20%231", false, literal_8420, Params{Param{Key: "user", Value: "slash%%%%2Fgordon"}, Param{Key: "project", Value: "Project%%%%20%231"}}},
	}, unescape)

	checkPriorities(t, tree)
}

func catchPanic(testFunc func()) (recv any) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}

type testRoute struct {
	path     string
	conflict bool
}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route.path, nil)
		})

		if route.conflict {
			if recv == nil {
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}
}

func TestTreeWildcardConflict(t *testing.T) {
	routes := []testRoute{
		{literal_8372, false},
		{literal_6792, false},
		{"/foo/bar", false},
		{"/foo/:name", false},
		{"/foo/:names", true},
		{"/cmd/*path", true},
		{"/cmd/:badvar", true},
		{"/cmd/:tool/names", false},
		{"/cmd/:tool/:badsub/details", true},
		{literal_2791, false},
		{"/src/:file", true},
		{"/src/static.json", true},
		{"/src/*filepathx", true},
		{"/src/", true},
		{"/src/foo/bar", true},
		{"/src1/", false},
		{"/src1/*filepath", true},
		{"/src2*filepath", true},
		{"/src2/*filepath", false},
		{literal_8590, false},
		{"/search/valid", false},
		{literal_0836, false},
		{"/user_x", false},
		{literal_0836, false},
		{"/id:id", false},
		{"/id/:id", false},
		{"/static/*file", false},
		{"/static/", true},
		{"/escape/test\\:d1", false},
		{"/escape/test\\:d2", false},
		{"/escape/test:param", false},
	}
	testRoutes(t, routes)
}

func TestCatchAllAfterSlash(t *testing.T) {
	routes := []testRoute{
		{"/non-leading-*catchall", true},
	}
	testRoutes(t, routes)
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		{literal_6792, false},
		{"/cmd/:tool", false},
		{literal_8372, false},
		{"/cmd/:tool/misc", false},
		{"/cmd/:tool/:othersub", true},
		{"/src/AUTHORS", false},
		{literal_2791, true},
		{"/user_x", false},
		{literal_0836, false},
		{"/id/:id", false},
		{"/id:id", false},
		{"/:id", false},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeDuplicatePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/doc/",
		literal_2791,
		literal_8590,
		literal_0836,
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf(literal_6458, route, recv)
		}

		// Add again
		recv = catchPanic(func() {
			tree.addRoute(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting duplicate route '%s", route)
		}
	}

	//printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/doc/", false, "/doc/", nil},
		{literal_3619, false, literal_2791, Params{Param{"filepath", "/some/file.png"}}},
		{literal_7309, false, literal_8590, Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, literal_0836, Params{Param{"name", "gopher"}}},
	})
}

func TestEmptyWildcardName(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/user:",
		"/user:/",
		"/cmd/:/",
		"/src/*",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting route with empty wildcard name '%s", route)
		}
	}
}

func TestTreeCatchAllConflict(t *testing.T) {
	routes := []testRoute{
		{"/src/*filepath/x", true},
		{"/src2/", false},
		{"/src2/*filepath/x", true},
		{"/src3/*filepath", false},
		{"/src3/*filepath/x", true},
	}
	testRoutes(t, routes)
}

func TestTreeCatchAllConflictRoot(t *testing.T) {
	routes := []testRoute{
		{"/", false},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeCatchMaxParams(t *testing.T) {
	tree := &node{}
	var route = "/cmd/*filepath"
	tree.addRoute(route, fakeHandler(route))
}

func TestTreeDoubleWildcard(t *testing.T) {
	const panicMsg = "only one wildcard per path segment is allowed"

	routes := [...]string{
		"/:foo:bar",
		"/:foo:bar/",
		"/:foo*bar",
	}

	for _, route := range routes {
		tree := &node{}
		recv := catchPanic(func() {
			tree.addRoute(route, nil)
		})

		if rs, ok := recv.(string); !ok || !strings.HasPrefix(rs, panicMsg) {
			t.Fatalf(`"Expected panic "%s" for route '%s', got "%v"`, panicMsg, route, recv)
		}
	}
}

/*func TestTreeDuplicateWildcard(t *testing.T) {
	tree := &node{}
	routes := [...]string{
		"/:id/:name/:id",
	}
	for _, route := range routes {
		...
	}
}*/

func TestTreeTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/b/",
		literal_8590,
		literal_0946,
		literal_2791,
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/admin",
		"/admin/:category",
		"/admin/:category/:page",
		"/doc",
		literal_2035,
		literal_0261,
		"/no/a",
		"/no/b",
		"/api/:page/:name",
		"/api/hello/:name/bar/",
		"/api/bar/:name",
		"/api/baz/foo",
		"/api/baz/foo/bar",
		"/blog/:p",
		"/posts/:b/:c",
		"/posts/b/:c/d/",
		"/vendor/:x/*y",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf(literal_6458, route, recv)
		}
	}

	tsrRoutes := [...]string{
		"/hi/",
		"/b",
		"/search/gopher/",
		literal_6792,
		"/src",
		"/x/",
		"/y",
		"/0/go/",
		"/1/go",
		"/a",
		"/admin/",
		"/admin/config/",
		"/admin/config/permissions/",
		"/doc/",
		"/admin/static/",
		"/admin/cfg/",
		"/admin/cfg/users/",
		"/api/hello/x/bar",
		"/api/baz/foo/",
		"/api/baz/bax/",
		"/api/bar/huh/",
		"/api/baz/foo/bar/",
		"/api/world/abc/",
		"/blog/pp/",
		"/posts/b/c/d",
		"/vendor/x",
	}

	for _, route := range tsrRoutes {
		value := tree.getValue(route, nil, getSkippedNodes(), false)
		if value.handlers != nil {
			t.Fatalf("non-nil handler for TSR route '%s", route)
		} else if !value.tsr {
			t.Errorf("expected TSR recommendation for route '%s'", route)
		}
	}

	noTsrRoutes := [...]string{
		"/",
		"/no",
		"/no/",
		"/_",
		"/_/",
		"/api",
		"/api/",
		"/api/hello/x/foo",
		"/api/baz/foo/bad",
		"/foo/p/p",
	}
	for _, route := range noTsrRoutes {
		value := tree.getValue(route, nil, getSkippedNodes(), false)
		if value.handlers != nil {
			t.Fatalf("non-nil handler for No-TSR route '%s", route)
		} else if value.tsr {
			t.Errorf("expected no TSR recommendation for route '%s'", route)
		}
	}
}

func TestTreeRootTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	recv := catchPanic(func() {
		tree.addRoute("/:test", fakeHandler("/:test"))
	})
	if recv != nil {
		t.Fatalf("panic inserting test route: %v", recv)
	}

	value := tree.getValue("/", nil, getSkippedNodes(), false)
	if value.handlers != nil {
		t.Fatalf("non-nil handler")
	} else if value.tsr {
		t.Errorf("expected no TSR recommendation")
	}
}

func TestRedirectTrailingSlash(t *testing.T) {
	var data = []struct {
		path string
	}{
		{"/hello/:name"},
		{"/hello/:name/123"},
		{"/hello/:name/234"},
	}

	node := &node{}
	for _, item := range data {
		node.addRoute(item.path, fakeHandler("test"))
	}

	value := node.getValue("/hello/abx/", nil, getSkippedNodes(), false)
	if value.tsr != true {
		t.Fatalf("want true, is false")
	}
}

func TestTreeFindCaseInsensitivePath(t *testing.T) {
	tree := &node{}

	longPath := "/l" + strings.Repeat("o", 128) + "ng"
	lOngPath := "/l" + strings.Repeat("O", 128) + "ng/"

	routes := [...]string{
		"/hi",
		"/b/",
		"/ABC/",
		literal_8590,
		literal_0946,
		literal_2791,
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/doc",
		literal_2035,
		literal_0261,
		"/doc/go/away",
		"/no/a",
		"/no/b",
		"/Π",
		"/u/apfêl/",
		literal_6471,
		literal_2864,
		literal_9028,
		literal_2981,
		"/w/♬",  // 3 byte
		"/w/♭/", // 3 byte, last byte differs
		"/w/𠜎",  // 4 byte
		"/w/𠜏/", // 4 byte
		longPath,
	}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf(literal_6458, route, recv)
		}
	}

	// Check out == in for all registered routes
	// With fixTrailingSlash = true
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, true)
		if !found {
			t.Errorf("Route '%s' not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route '%s': %s", route, string(out))
		}
	}
	// With fixTrailingSlash = false
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, false)
		if !found {
			t.Errorf("Route '%s' not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route '%s': %s", route, string(out))
		}
	}

	tests := []struct {
		in    string
		out   string
		found bool
		slash bool
	}{
		{"/HI", "/hi", true, false},
		{"/HI/", "/hi", true, true},
		{"/B", "/b/", true, true},
		{"/B/", "/b/", true, false},
		{"/abc", "/ABC/", true, true},
		{"/abc/", "/ABC/", true, false},
		{"/aBc", "/ABC/", true, true},
		{"/aBc/", "/ABC/", true, false},
		{"/abC", "/ABC/", true, true},
		{"/abC/", "/ABC/", true, false},
		{"/SEARCH/QUERY", "/search/QUERY", true, false},
		{"/SEARCH/QUERY/", "/search/QUERY", true, true},
		{"/CMD/TOOL/", "/cmd/TOOL/", true, false},
		{"/CMD/TOOL", "/cmd/TOOL/", true, true},
		{"/SRC/FILE/PATH", "/src/FILE/PATH", true, false},
		{"/x/Y", "/x/y", true, false},
		{"/x/Y/", "/x/y", true, true},
		{"/X/y", "/x/y", true, false},
		{"/X/y/", "/x/y", true, true},
		{"/X/Y", "/x/y", true, false},
		{"/X/Y/", "/x/y", true, true},
		{"/Y/", "/y/", true, false},
		{"/Y", "/y/", true, true},
		{"/Y/z", "/y/z", true, false},
		{"/Y/z/", "/y/z", true, true},
		{"/Y/Z", "/y/z", true, false},
		{"/Y/Z/", "/y/z", true, true},
		{"/y/Z", "/y/z", true, false},
		{"/y/Z/", "/y/z", true, true},
		{"/Aa", "/aa", true, false},
		{"/Aa/", "/aa", true, true},
		{"/AA", "/aa", true, false},
		{"/AA/", "/aa", true, true},
		{"/aA", "/aa", true, false},
		{"/aA/", "/aa", true, true},
		{"/A/", "/a/", true, false},
		{"/A", "/a/", true, true},
		{"/DOC", "/doc", true, false},
		{"/DOC/", "/doc", true, true},
		{"/NO", "", false, true},
		{"/DOC/GO", "", false, true},
		{"/π", "/Π", true, false},
		{"/π/", "/Π", true, true},
		{"/u/ÄPFÊL/", literal_6471, true, false},
		{"/u/ÄPFÊL", literal_6471, true, true},
		{"/u/ÖPFÊL/", literal_2864, true, true},
		{"/u/ÖPFÊL", literal_2864, true, false},
		{"/v/äpfêL/", literal_9028, true, false},
		{"/v/äpfêL", literal_9028, true, true},
		{"/v/öpfêL/", literal_2981, true, true},
		{"/v/öpfêL", literal_2981, true, false},
		{"/w/♬/", "/w/♬", true, true},
		{"/w/♭", "/w/♭/", true, true},
		{"/w/𠜎/", "/w/𠜎", true, true},
		{"/w/𠜏", "/w/𠜏/", true, true},
		{lOngPath, longPath, true, true},
	}
	// With fixTrailingSlash = true
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, true)
		if found != test.found || (found && (string(out) != test.out)) {
			t.Errorf("Wrong result for '%s': got %s, %t; want %s, %t",
				test.in, string(out), found, test.out, test.found)
			return
		}
	}
	// With fixTrailingSlash = false
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, false)
		if test.slash {
			if found { // test needs a trailingSlash fix. It must not be found!
				t.Errorf("Found without fixTrailingSlash: %s; got %s", test.in, string(out))
			}
		} else {
			if found != test.found || (found && (string(out) != test.out)) {
				t.Errorf("Wrong result for '%s': got %s, %t; want %s, %t",
					test.in, string(out), found, test.out, test.found)
				return
			}
		}
	}
}

func TestTreeInvalidNodeType(t *testing.T) {
	const panicMsg = "invalid node type"

	tree := &node{}
	tree.addRoute("/", fakeHandler("/"))
	tree.addRoute("/:page", fakeHandler("/:page"))

	// set invalid node type
	tree.children[0].nType = 42

	// normal lookup
	recv := catchPanic(func() {
		tree.getValue("/test", nil, getSkippedNodes(), false)
	})
	if rs, ok := recv.(string); !ok || rs != panicMsg {
		t.Fatalf("Expected panic '"+panicMsg+"', got '%v'", recv)
	}

	// case-insensitive lookup
	recv = catchPanic(func() {
		tree.findCaseInsensitivePath("/test", true)
	})
	if rs, ok := recv.(string); !ok || rs != panicMsg {
		t.Fatalf("Expected panic '"+panicMsg+"', got '%v'", recv)
	}
}

func TestTreeInvalidParamsType(t *testing.T) {
	tree := &node{}
	// add a child with wildcard
	route := "/:path"
	tree.addRoute(route, fakeHandler(route))

	// set invalid Params type
	params := make(Params, 0)

	// try to trigger slice bounds out of range with capacity 0
	tree.getValue("/test", &params, getSkippedNodes(), false)
}

func TestTreeExpandParamsCapacity(t *testing.T) {
	data := []struct {
		path string
	}{
		{"/:path"},
		{"/*path"},
	}

	for _, item := range data {
		tree := &node{}
		tree.addRoute(item.path, fakeHandler(item.path))
		params := make(Params, 0)

		value := tree.getValue("/test", &params, getSkippedNodes(), false)

		if value.params == nil {
			t.Errorf("Expected %s params to be set, but they weren't", item.path)
			continue
		}

		if len(*value.params) != 1 {
			t.Errorf("Wrong number of %s params: got %d, want %d",
				item.path, len(*value.params), 1)
			continue
		}
	}
}

func TestTreeWildcardConflictEx(t *testing.T) {
	conflicts := [...]struct {
		route        string
		segPath      string
		existPath    string
		existSegPath string
	}{
		{"/who/are/foo", "/foo", `/who/are/\*you`, `/\*you`},
		{"/who/are/foo/", "/foo/", `/who/are/\*you`, `/\*you`},
		{"/who/are/foo/bar", "/foo/bar", `/who/are/\*you`, `/\*you`},
		{"/con:nection", ":nection", `/con:tact`, `:tact`},
	}

	for _, conflict := range conflicts {
		// I have to re-create a 'tree', because the 'tree' will be
		// in an inconsistent state when the loop recovers from the
		// panic which threw by 'addRoute' function.
		tree := &node{}
		routes := [...]string{
			"/con:tact",
			"/who/are/*you",
			"/who/foo/hello",
		}

		for _, route := range routes {
			tree.addRoute(route, fakeHandler(route))
		}

		recv := catchPanic(func() {
			tree.addRoute(conflict.route, fakeHandler(conflict.route))
		})

		if !regexp.MustCompile(fmt.Sprintf("'%s' in new path .* conflicts with existing wildcard '%s' in existing prefix '%s'", conflict.segPath, conflict.existSegPath, conflict.existPath)).MatchString(fmt.Sprint(recv)) {
			t.Fatalf("invalid wildcard conflict error (%v)", recv)
		}
	}
}

func TestTreeInvalidEscape(t *testing.T) {
	routes := map[string]bool{
		"/r1/r":    true,
		"/r2/:r":   true,
		"/r3/\\:r": true,
	}
	tree := &node{}
	for route, valid := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv == nil != valid {
			t.Fatalf("%s should be %t but got %v", route, valid, recv)
		}
	}
}

const literal_4217 = "/contact"

const literal_2035 = "/doc/go_faq.html"

const literal_0261 = "/doc/go1.html"

const literal_0946 = "/cmd/:tool/"

const literal_8372 = "/cmd/:tool/:sub"

const literal_5389 = "/cmd/whoami"

const literal_6049 = "/cmd/whoami/root"

const literal_4827 = "/cmd/whoami/root/"

const literal_2791 = "/src/*filepath"

const literal_1489 = "/search/"

const literal_8590 = "/search/:query"

const literal_8605 = "/search/gin-gonic"

const literal_0652 = "/search/google"

const literal_0836 = "/user_:name"

const literal_6470 = "/files/:dir/*filepath"

const literal_8420 = "/info/:user/project/:project"

const literal_2845 = "/:cc/cc"

const literal_5382 = "/:cc/:dd/ee"

const literal_0384 = "/:cc/:dd/:ee/:ff/:gg/hh"

const literal_2645 = "/get/test/abc/"

const literal_0297 = "/get/:param/abc/"

const literal_4176 = "/something/:paramname/thirdthing"

const literal_4691 = "/something/secondthing/test"

const literal_5986 = "/get/abc"

const literal_0974 = "/get/:param"

const literal_8120 = "/get/abc/123abc"

const literal_9526 = "/get/abc/:param"

const literal_3065 = "/get/abc/123abc/xxx8"

const literal_3026 = "/get/abc/123abc/:param"

const literal_6074 = "/get/abc/123abc/xxx8/1234"

const literal_4320 = "/get/abc/123abc/xxx8/:param"

const literal_9802 = "/get/abc/123abc/xxx8/1234/ffas"

const literal_0125 = "/get/abc/123abc/xxx8/1234/:param"

const literal_6970 = "/get/abc/123abc/xxx8/1234/kkdd/12c"

const literal_6843 = "/get/abc/123abc/xxx8/1234/kkdd/:param"

const literal_9516 = "/get/abc/:param/test"

const literal_3654 = "/get/abc/escaped_colon/test\\:param"

const literal_3619 = "/src/some/file.png"

const literal_7309 = "/search/someth!ng+in+ünìcodé"

const literal_4578 = "/info/:user"

const literal_6792 = "/cmd/vet"

const literal_6458 = "panic inserting route '%s': %v"

const literal_6471 = "/u/äpfêl/"

const literal_2864 = "/u/öpfêl"

const literal_9028 = "/v/Äpfêl/"

const literal_2981 = "/v/Öpfêl"
