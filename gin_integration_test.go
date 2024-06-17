// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// params[0]=url example:http://127.0.0.1:8080/index (cannot be empty)
// params[1]=response status (custom compare status) default:"200 OK"
// params[2]=response body (custom compare content)  default:literal_7812
func testRequest(t *testing.T, params ...string) {

	if len(params) == 0 {
		t.Fatal("url cannot be empty")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(params[0])
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, ioerr := io.ReadAll(resp.Body)
	assert.NoError(t, ioerr)

	var responseStatus = "200 OK"
	if len(params) > 1 && params[1] != "" {
		responseStatus = params[1]
	}

	var responseBody = literal_7812
	if len(params) > 2 && params[2] != "" {
		responseBody = params[2]
	}

	assert.Equal(t, responseStatus, resp.Status, "should get a "+responseStatus)
	if responseStatus == "200 OK" {
		assert.Equal(t, responseBody, string(body), literal_3982)
	}
}

func TestRunEmpty(t *testing.T) {
	os.Setenv("PORT", "")
	router := New()
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.Run())
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.Run(":8080"))
	testRequest(t, "http://localhost:8080/example")
}

func TestBadTrustedCIDRs(t *testing.T) {
	router := New()
	assert.Error(t, router.SetTrustedProxies([]string{"hello/world"}))
}

/* legacy tests
func TestBadTrustedCIDRsForRun(t *testing.T) {
	os.Setenv("PORT", "")
	router := New()
	router.TrustedProxies = []string{"hello/world"}
	assert.Error(t, router.Run(":8080"))
}

func TestBadTrustedCIDRsForRunUnix(t *testing.T) {
	router := New()
	router.TrustedProxies = []string{"hello/world"}

	unixTestSocket := filepath.Join(os.TempDir(), "unix_unit_test")

	defer os.Remove(unixTestSocket)

	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.Error(t, router.RunUnix(unixTestSocket))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)
}

func TestBadTrustedCIDRsForRunFd(t *testing.T) {
	router := New()
	router.TrustedProxies = []string{"hello/world"}

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)
	socketFile, err := listener.File()
	assert.NoError(t, err)

	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.Error(t, router.RunFd(int(socketFile.Fd())))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)
}

func TestBadTrustedCIDRsForRunListener(t *testing.T) {
	router := New()
	router.TrustedProxies = []string{"hello/world"}

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.Error(t, router.RunListener(listener))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)
}

func TestBadTrustedCIDRsForRunTLS(t *testing.T) {
	os.Setenv("PORT", "")
	router := New()
	router.TrustedProxies = []string{"hello/world"}
	assert.Error(t, router.RunTLS(":8080", literal_8762, literal_9713))
}
*/

func TestRunTLS(t *testing.T) {
	router := New()
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })

		assert.NoError(t, router.RunTLS(":8443", literal_8762, literal_9713))
	}()

	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.RunTLS(":8443", literal_8762, literal_9713))
	testRequest(t, "https://localhost:8443/example")
}

func TestPusher(t *testing.T) {
	var html = template.Must(template.New("https").Parse(`
<html>
<head>
  <title>Https Test</title>
  <script src="/assets/app.js"></script>
</head>
<body>
  <h1 style="color:red;">Welcome, Ginner!</h1>
</body>
</html>
`))

	router := New()
	router.Static("./assets", "./assets")
	router.SetHTMLTemplate(html)

	go func() {
		router.GET("/pusher", func(c *Context) {
			if pusher := c.Writer.Pusher(); pusher != nil {
				err := pusher.Push("/assets/app.js", nil)
				assert.NoError(t, err)
			}
			c.String(http.StatusOK, literal_7812)
		})

		assert.NoError(t, router.RunTLS(":8449", literal_8762, literal_9713))
	}()

	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.RunTLS(":8449", literal_8762, literal_9713))
	testRequest(t, "https://localhost:8449/pusher")
}

func TestRunEmptyWithEnv(t *testing.T) {
	os.Setenv("PORT", "3123")
	router := New()
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.Run())
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.Run(":3123"))
	testRequest(t, "http://localhost:3123/example")
}

func TestRunTooMuchParams(t *testing.T) {
	router := New()
	assert.Panics(t, func() {
		assert.NoError(t, router.Run("2", "2"))
	})
}

func TestRunWithPort(t *testing.T) {
	router := New()
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.Run(":5150"))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.Run(":5150"))
	testRequest(t, "http://localhost:5150/example")
}

func TestUnixSocket(t *testing.T) {
	router := New()

	unixTestSocket := filepath.Join(os.TempDir(), "unix_unit_test")

	defer os.Remove(unixTestSocket)

	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.RunUnix(unixTestSocket))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	c, err := net.Dial("unix", unixTestSocket)
	assert.NoError(t, err)

	fmt.Fprint(c, literal_0968)
	scanner := bufio.NewScanner(c)
	var response string
	for scanner.Scan() {
		response += scanner.Text()
	}
	assert.Contains(t, response, literal_0634, literal_6914)
	assert.Contains(t, response, literal_7812, literal_3982)
}

func TestBadUnixSocket(t *testing.T) {
	router := New()
	assert.Error(t, router.RunUnix("#/tmp/unix_unit_test"))
}

func TestRunQUIC(t *testing.T) {
	router := New()
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })

		assert.NoError(t, router.RunQUIC(":8443", literal_8762, literal_9713))
	}()

	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	assert.Error(t, router.RunQUIC(":8443", literal_8762, literal_9713))
	testRequest(t, "https://localhost:8443/example")
}

func TestFileDescriptor(t *testing.T) {
	router := New()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)
	socketFile, err := listener.File()
	if isWindows() {
		// not supported by windows, it is unimplemented now
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}

	if socketFile == nil {
		return
	}

	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.RunFd(int(socketFile.Fd())))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	c, err := net.Dial("tcp", listener.Addr().String())
	assert.NoError(t, err)

	fmt.Fprintf(c, literal_0968)
	scanner := bufio.NewScanner(c)
	var response string
	for scanner.Scan() {
		response += scanner.Text()
	}
	assert.Contains(t, response, literal_0634, literal_6914)
	assert.Contains(t, response, literal_7812, literal_3982)
}

func TestBadFileDescriptor(t *testing.T) {
	router := New()
	assert.Error(t, router.RunFd(0))
}

func TestListener(t *testing.T) {
	router := New()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)
	go func() {
		router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })
		assert.NoError(t, router.RunListener(listener))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	c, err := net.Dial("tcp", listener.Addr().String())
	assert.NoError(t, err)

	fmt.Fprintf(c, literal_0968)
	scanner := bufio.NewScanner(c)
	var response string
	for scanner.Scan() {
		response += scanner.Text()
	}
	assert.Contains(t, response, literal_0634, literal_6914)
	assert.Contains(t, response, literal_7812, literal_3982)
}

func TestBadListener(t *testing.T) {
	router := New()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:10086")
	assert.NoError(t, err)
	listener, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)
	listener.Close()
	assert.Error(t, router.RunListener(listener))
}

func TestWithHttptestWithAutoSelectedPort(t *testing.T) {
	router := New()
	router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })

	ts := httptest.NewServer(router)
	defer ts.Close()

	testRequest(t, ts.URL+literal_3274)
}

func TestConcurrentHandleContext(t *testing.T) {
	router := New()
	router.GET("/", func(c *Context) {
		c.Request.URL.Path = literal_3274
		router.HandleContext(c)
	})
	router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })

	var wg sync.WaitGroup
	iterations := 200
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			testGetRequestHandler(t, router, "/")
			wg.Done()
		}()
	}
	wg.Wait()
}

// func TestWithHttptestWithSpecifiedPort(t *testing.T) {
// 	router := New()
// 	router.GET(literal_3274, func(c *Context) { c.String(http.StatusOK, literal_7812) })

// 	l, _ := net.Listen("tcp", ":8033")
// 	ts := httptest.Server{
// 		Listener: l,
// 		Config:   &http.Server{Handler: router},
// 	}
// 	ts.Start()
// 	defer ts.Close()

// 	testRequest(t, "http://localhost:8033/example")
// }

func testGetRequestHandler(t *testing.T, h http.Handler, url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, literal_7812, w.Body.String(), literal_3982)
	assert.Equal(t, 200, w.Code, literal_6914)
}

func TestTreeRunDynamicRouting(t *testing.T) {
	router := New()
	router.GET(literal_6410, func(c *Context) { c.String(http.StatusOK, literal_6410) })
	router.GET(literal_7534, func(c *Context) { c.String(http.StatusOK, literal_7534) })
	router.GET("/", func(c *Context) { c.String(http.StatusOK, "home") })
	router.GET("/:cc", func(c *Context) { c.String(http.StatusOK, "/:cc") })
	router.GET(literal_5123, func(c *Context) { c.String(http.StatusOK, literal_5123) })
	router.GET(literal_4685, func(c *Context) { c.String(http.StatusOK, literal_4685) })
	router.GET("/c1/:dd/f1", func(c *Context) { c.String(http.StatusOK, "/c1/:dd/f1") })
	router.GET("/c1/:dd/f2", func(c *Context) { c.String(http.StatusOK, "/c1/:dd/f2") })
	router.GET(literal_5423, func(c *Context) { c.String(http.StatusOK, literal_5423) })
	router.GET(literal_9834, func(c *Context) { c.String(http.StatusOK, literal_9834) })
	router.GET(literal_28450, func(c *Context) { c.String(http.StatusOK, literal_28450) })
	router.GET(literal_4351, func(c *Context) { c.String(http.StatusOK, literal_4351) })
	router.GET(literal_0629, func(c *Context) { c.String(http.StatusOK, literal_0629) })
	router.GET(literal_6203, func(c *Context) { c.String(http.StatusOK, literal_6203) })
	router.GET(literal_0839, func(c *Context) { c.String(http.StatusOK, literal_0839) })
	router.GET(literal_5286, func(c *Context) { c.String(http.StatusOK, literal_5286) })
	router.GET(literal_1728, func(c *Context) { c.String(http.StatusOK, literal_1728) })
	router.GET(literal_8924, func(c *Context) { c.String(http.StatusOK, literal_8924) })
	router.GET(literal_2708, func(c *Context) { c.String(http.StatusOK, literal_2708) })
	router.GET(literal_1235, func(c *Context) { c.String(http.StatusOK, literal_1235) })
	router.GET(literal_2158, func(c *Context) { c.String(http.StatusOK, literal_2158) })
	router.GET(literal_7609, func(c *Context) { c.String(http.StatusOK, literal_7609) })
	router.GET(literal_2451, func(c *Context) { c.String(http.StatusOK, literal_2451) })
	router.GET(literal_9168, func(c *Context) { c.String(http.StatusOK, literal_9168) })
	router.GET(literal_1089, func(c *Context) { c.String(http.StatusOK, literal_1089) })
	router.GET(literal_2471, func(c *Context) { c.String(http.StatusOK, literal_2471) })
	router.GET(literal_0271, func(c *Context) { c.String(http.StatusOK, literal_0271) })
	router.GET(literal_6743, func(c *Context) { c.String(http.StatusOK, literal_6743) })
	router.GET(literal_7405, func(c *Context) { c.String(http.StatusOK, literal_7405) })
	router.GET(literal_3928, func(c *Context) { c.String(http.StatusOK, literal_3928) })
	router.GET(literal_5719, func(c *Context) { c.String(http.StatusOK, literal_5719) })
	router.GET(literal_9425, func(c *Context) { c.String(http.StatusOK, literal_9425) })
	router.GET(literal_4860, func(c *Context) { c.String(http.StatusOK, literal_4860) })
	router.GET(literal_7139, func(c *Context) { c.String(http.StatusOK, literal_7139) })
	router.GET(literal_1954, func(c *Context) { c.String(http.StatusOK, literal_1954) })
	router.GET(literal_3095, func(c *Context) { c.String(http.StatusOK, literal_3095) })
	router.GET(literal_8617, func(c *Context) { c.String(http.StatusOK, literal_8617) })

	ts := httptest.NewServer(router)
	defer ts.Close()

	testRequest(t, ts.URL+"/", "", "home")
	testRequest(t, ts.URL+"/aa/aa", "", literal_6410)
	testRequest(t, ts.URL+"/ab/ab", "", literal_7534)
	testRequest(t, ts.URL+"/all", "", "/:cc")
	testRequest(t, ts.URL+"/all/cc", "", literal_5423)
	testRequest(t, ts.URL+"/a/cc", "", literal_5423)
	testRequest(t, ts.URL+"/c1/d/e", "", literal_5123)
	testRequest(t, ts.URL+"/c1/d/e1", "", literal_4685)
	testRequest(t, ts.URL+"/c1/d/ee", "", literal_9834)
	testRequest(t, ts.URL+"/c1/d/f", "", literal_28450)
	testRequest(t, ts.URL+"/c/d/ee", "", literal_9834)
	testRequest(t, ts.URL+"/c/d/e/ff", "", literal_4351)
	testRequest(t, ts.URL+"/c/d/e/f/gg", "", literal_0629)
	testRequest(t, ts.URL+"/c/d/e/f/g/hh", "", literal_6203)
	testRequest(t, ts.URL+"/cc/dd/ee/ff/gg/hh", "", literal_6203)
	testRequest(t, ts.URL+"/a", "", "/:cc")
	testRequest(t, ts.URL+"/d", "", "/:cc")
	testRequest(t, ts.URL+"/ad", "", "/:cc")
	testRequest(t, ts.URL+"/dd", "", "/:cc")
	testRequest(t, ts.URL+"/aa", "", "/:cc")
	testRequest(t, ts.URL+"/aaa", "", "/:cc")
	testRequest(t, ts.URL+"/aaa/cc", "", literal_5423)
	testRequest(t, ts.URL+"/ab", "", "/:cc")
	testRequest(t, ts.URL+"/abb", "", "/:cc")
	testRequest(t, ts.URL+"/abb/cc", "", literal_5423)
	testRequest(t, ts.URL+"/dddaa", "", "/:cc")
	testRequest(t, ts.URL+"/allxxxx", "", "/:cc")
	testRequest(t, ts.URL+"/alldd", "", "/:cc")
	testRequest(t, ts.URL+"/cc/cc", "", literal_5423)
	testRequest(t, ts.URL+"/ccc/cc", "", literal_5423)
	testRequest(t, ts.URL+"/deedwjfs/cc", "", literal_5423)
	testRequest(t, ts.URL+"/acllcc/cc", "", literal_5423)
	testRequest(t, ts.URL+literal_0839, "", literal_0839)
	testRequest(t, ts.URL+"/get/testaa/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/te/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/xx/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/tt/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/a/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/t/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/aa/abc/", "", literal_5286)
	testRequest(t, ts.URL+"/get/abas/abc/", "", literal_5286)
	testRequest(t, ts.URL+literal_8924, "", literal_8924)
	testRequest(t, ts.URL+"/something/secondthingaaaa/thirdthing", "", literal_1728)
	testRequest(t, ts.URL+"/something/abcdad/thirdthing", "", literal_1728)
	testRequest(t, ts.URL+"/something/se/thirdthing", "", literal_1728)
	testRequest(t, ts.URL+"/something/s/thirdthing", "", literal_1728)
	testRequest(t, ts.URL+"/something/secondthing/thirdthing", "", literal_1728)
	testRequest(t, ts.URL+literal_2708, "", literal_2708)
	testRequest(t, ts.URL+"/get/a", "", literal_1235)
	testRequest(t, ts.URL+"/get/abz", "", literal_1235)
	testRequest(t, ts.URL+"/get/12a", "", literal_1235)
	testRequest(t, ts.URL+"/get/abcd", "", literal_1235)
	testRequest(t, ts.URL+literal_2158, "", literal_2158)
	testRequest(t, ts.URL+"/get/abc/12", "", literal_7609)
	testRequest(t, ts.URL+"/get/abc/123ab", "", literal_7609)
	testRequest(t, ts.URL+"/get/abc/xyz", "", literal_7609)
	testRequest(t, ts.URL+"/get/abc/123abcddxx", "", literal_7609)
	testRequest(t, ts.URL+literal_2451, "", literal_2451)
	testRequest(t, ts.URL+"/get/abc/123abc/x", "", literal_9168)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx", "", literal_9168)
	testRequest(t, ts.URL+"/get/abc/123abc/abc", "", literal_9168)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8xxas", "", literal_9168)
	testRequest(t, ts.URL+literal_1089, "", literal_1089)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1", "", literal_2471)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/123", "", literal_2471)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/78k", "", literal_2471)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234xxxd", "", literal_2471)
	testRequest(t, ts.URL+literal_0271, "", literal_0271)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/f", "", literal_6743)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/ffa", "", literal_6743)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kka", "", literal_6743)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/ffas321", "", literal_6743)
	testRequest(t, ts.URL+literal_7405, "", literal_7405)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kkdd/1", "", literal_3928)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kkdd/12", "", literal_3928)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kkdd/12b", "", literal_3928)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kkdd/34", "", literal_3928)
	testRequest(t, ts.URL+"/get/abc/123abc/xxx8/1234/kkdd/12c2e3", "", literal_3928)
	testRequest(t, ts.URL+"/get/abc/12/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abdd/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abdddf/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123ab/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abgg/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abff/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abffff/test", "", literal_5719)
	testRequest(t, ts.URL+"/get/abc/123abd/test", "", literal_9425)
	testRequest(t, ts.URL+"/get/abc/123abddd/test", "", literal_4860)
	testRequest(t, ts.URL+"/get/abc/123/test22", "", literal_7139)
	testRequest(t, ts.URL+"/get/abc/123abg/test", "", literal_1954)
	testRequest(t, ts.URL+"/get/abc/123abf/testss", "", literal_3095)
	testRequest(t, ts.URL+"/get/abc/123abfff/te", "", literal_8617)
	// 404 not found
	testRequest(t, ts.URL+"/c/d/e", literal_9815)
	testRequest(t, ts.URL+"/c/d/e1", literal_9815)
	testRequest(t, ts.URL+"/c/d/eee", literal_9815)
	testRequest(t, ts.URL+"/c1/d/eee", literal_9815)
	testRequest(t, ts.URL+"/c1/d/e2", literal_9815)
	testRequest(t, ts.URL+"/cc/dd/ee/ff/gg/hh1", literal_9815)
	testRequest(t, ts.URL+"/a/dd", literal_9815)
	testRequest(t, ts.URL+"/addr/dd/aa", literal_9815)
	testRequest(t, ts.URL+"/something/secondthing/121", literal_9815)
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func TestEscapedColon(t *testing.T) {
	router := New()
	f := func(u string) {
		router.GET(u, func(c *Context) { c.String(http.StatusOK, u) })
	}
	f("/r/r\\:r")
	f(literal_6310)
	f(literal_2638)
	f("/r/r/\\:r")
	f("/r/r/r\\:r")
	assert.Panics(t, func() {
		f("\\foo:")
	})

	router.updateRouteTrees()
	ts := httptest.NewServer(router)
	defer ts.Close()

	testRequest(t, ts.URL+"/r/r123", "", literal_6310)
	testRequest(t, ts.URL+literal_6310, "", "/r/r\\:r")
	testRequest(t, ts.URL+"/r/r/r123", "", literal_2638)
	testRequest(t, ts.URL+literal_2638, "", "/r/r/\\:r")
	testRequest(t, ts.URL+"/r/r/r:r", "", "/r/r/r\\:r")
}

const literal_7812 = "it worked"

const literal_3982 = "resp body should match"

const literal_3274 = "/example"

const literal_8762 = "./testdata/certificate/cert.pem"

const literal_9713 = "./testdata/certificate/key.pem"

const literal_0968 = "GET /example HTTP/1.0\r\n\r\n"

const literal_0634 = "HTTP/1.0 200"

const literal_6914 = "should get a 200"

const literal_6410 = "/aa/*xx"

const literal_7534 = "/ab/*xx"

const literal_5123 = "/c1/:dd/e"

const literal_4685 = "/c1/:dd/e1"

const literal_5423 = "/:cc/cc"

const literal_9834 = "/:cc/:dd/ee"

const literal_28450 = "/:cc/:dd/f"

const literal_4351 = "/:cc/:dd/:ee/ff"

const literal_0629 = "/:cc/:dd/:ee/:ff/gg"

const literal_6203 = "/:cc/:dd/:ee/:ff/:gg/hh"

const literal_0839 = "/get/test/abc/"

const literal_5286 = "/get/:param/abc/"

const literal_1728 = "/something/:paramname/thirdthing"

const literal_8924 = "/something/secondthing/test"

const literal_2708 = "/get/abc"

const literal_1235 = "/get/:param"

const literal_2158 = "/get/abc/123abc"

const literal_7609 = "/get/abc/:param"

const literal_2451 = "/get/abc/123abc/xxx8"

const literal_9168 = "/get/abc/123abc/:param"

const literal_1089 = "/get/abc/123abc/xxx8/1234"

const literal_2471 = "/get/abc/123abc/xxx8/:param"

const literal_0271 = "/get/abc/123abc/xxx8/1234/ffas"

const literal_6743 = "/get/abc/123abc/xxx8/1234/:param"

const literal_7405 = "/get/abc/123abc/xxx8/1234/kkdd/12c"

const literal_3928 = "/get/abc/123abc/xxx8/1234/kkdd/:param"

const literal_5719 = "/get/abc/:param/test"

const literal_9425 = "/get/abc/123abd/:param"

const literal_4860 = "/get/abc/123abddd/:param"

const literal_7139 = "/get/abc/123/:param"

const literal_1954 = "/get/abc/123abg/:param"

const literal_3095 = "/get/abc/123abf/:param"

const literal_8617 = "/get/abc/123abfff/:param"

const literal_9815 = "404 Not Found"

const literal_6310 = "/r/r:r"

const literal_2638 = "/r/r/:r"
