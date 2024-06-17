// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"encoding/xml"
	"errors"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jialequ/mpgw/internal/json"
	testdata "github.com/jialequ/mpgw/testdata/protoexample"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// test errors

func TestRenderJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}

	(JSON{data}).WriteContentType(w)
	assert.Equal(t, literal_3516, w.Header().Get(literal_2953))

	err := (JSON{data}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\",\"html\":\"\\u003cb\\u003e\"}", w.Body.String())
	assert.Equal(t, literal_3516, w.Header().Get(literal_2953))
}

func TestRenderJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	assert.Error(t, (JSON{data}).Render(w))
}

func TestRenderIndentedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
		"bar": "foo",
	}

	err := (IndentedJSON{data}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "{\n    \"bar\": \"foo\",\n    \"foo\": \"bar\"\n}", w.Body.String())
	assert.Equal(t, literal_3516, w.Header().Get(literal_2953))
}

func TestRenderIndentedJSONPanics(t *testing.T) {
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (IndentedJSON{data}).Render(w)
	assert.Error(t, err)
}

func TestRenderSecureJSON(t *testing.T) {
	w1 := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(SecureJSON{literal_9015, data}).WriteContentType(w1)
	assert.Equal(t, literal_3516, w1.Header().Get(literal_2953))

	err1 := (SecureJSON{literal_9015, data}).Render(w1)

	assert.NoError(t, err1)
	assert.Equal(t, "{\"foo\":\"bar\"}", w1.Body.String())
	assert.Equal(t, literal_3516, w1.Header().Get(literal_2953))

	w2 := httptest.NewRecorder()
	datas := []map[string]any{{
		"foo": "bar",
	}, {
		"bar": "foo",
	}}

	err2 := (SecureJSON{literal_9015, datas}).Render(w2)
	assert.NoError(t, err2)
	assert.Equal(t, "while(1);[{\"foo\":\"bar\"},{\"bar\":\"foo\"}]", w2.Body.String())
	assert.Equal(t, literal_3516, w2.Header().Get(literal_2953))
}

func TestRenderSecureJSONFail(t *testing.T) {
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (SecureJSON{literal_9015, data}).Render(w)
	assert.Error(t, err)
}

func TestRenderJsonpJSON(t *testing.T) {
	w1 := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(JsonpJSON{"x", data}).WriteContentType(w1)
	assert.Equal(t, literal_8096, w1.Header().Get(literal_2953))

	err1 := (JsonpJSON{"x", data}).Render(w1)

	assert.NoError(t, err1)
	assert.Equal(t, "x({\"foo\":\"bar\"});", w1.Body.String())
	assert.Equal(t, literal_8096, w1.Header().Get(literal_2953))

	w2 := httptest.NewRecorder()
	datas := []map[string]any{{
		"foo": "bar",
	}, {
		"bar": "foo",
	}}

	err2 := (JsonpJSON{"x", datas}).Render(w2)
	assert.NoError(t, err2)
	assert.Equal(t, "x([{\"foo\":\"bar\"},{\"bar\":\"foo\"}]);", w2.Body.String())
	assert.Equal(t, literal_8096, w2.Header().Get(literal_2953))
}

type errorWriter struct {
	bufString string
	*httptest.ResponseRecorder
}

var _ http.ResponseWriter = (*errorWriter)(nil)

func (w *errorWriter) Write(buf []byte) (int, error) {
	if string(buf) == w.bufString {
		return 0, errors.New(`write "` + w.bufString + `" error`)
	}
	return w.ResponseRecorder.Write(buf)
}

func TestRenderJsonpJSONError(t *testing.T) {
	ew := &errorWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}

	jsonpJSON := JsonpJSON{
		Callback: "foo",
		Data: map[string]string{
			"foo": "bar",
		},
	}

	cb := template.JSEscapeString(jsonpJSON.Callback)
	ew.bufString = cb
	err := jsonpJSON.Render(ew) // error was returned while writing callback
	assert.Equal(t, `write "`+cb+`" error`, err.Error())

	ew.bufString = `(`
	err = jsonpJSON.Render(ew)
	assert.Equal(t, `write "`+`(`+`" error`, err.Error())

	data, _ := json.Marshal(jsonpJSON.Data) // error was returned while writing data
	ew.bufString = string(data)
	err = jsonpJSON.Render(ew)
	assert.Equal(t, `write "`+string(data)+`" error`, err.Error())

	ew.bufString = `);`
	err = jsonpJSON.Render(ew)
	assert.Equal(t, `write "`+`);`+`" error`, err.Error())
}

func TestRenderJsonpJSONError2(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}
	(JsonpJSON{"", data}).WriteContentType(w)
	assert.Equal(t, literal_8096, w.Header().Get(literal_2953))

	e := (JsonpJSON{"", data}).Render(w)
	assert.NoError(t, e)

	assert.Equal(t, "{\"foo\":\"bar\"}", w.Body.String())
	assert.Equal(t, literal_8096, w.Header().Get(literal_2953))
}

func TestRenderJsonpJSONFail(t *testing.T) {
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (JsonpJSON{"x", data}).Render(w)
	assert.Error(t, err)
}

func TestRenderAsciiJSON(t *testing.T) {
	w1 := httptest.NewRecorder()
	data1 := map[string]any{
		"lang": "GO语言",
		"tag":  "<br>",
	}

	err := (AsciiJSON{data1}).Render(w1)

	assert.NoError(t, err)
	assert.Equal(t, "{\"lang\":\"GO\\u8bed\\u8a00\",\"tag\":\"\\u003cbr\\u003e\"}", w1.Body.String())
	assert.Equal(t, "application/json", w1.Header().Get(literal_2953))

	w2 := httptest.NewRecorder()
	data2 := 3.1415926

	err = (AsciiJSON{data2}).Render(w2)
	assert.NoError(t, err)
	assert.Equal(t, "3.1415926", w2.Body.String())
}

func TestRenderAsciiJSONFail(t *testing.T) {
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	assert.Error(t, (AsciiJSON{data}).Render(w))
}

func TestRenderPureJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}
	err := (PureJSON{data}).Render(w)
	assert.NoError(t, err)
	assert.Equal(t, "{\"foo\":\"bar\",\"html\":\"<b>\"}\n", w.Body.String())
	assert.Equal(t, literal_3516, w.Header().Get(literal_2953))
}

type xmlmap map[string]any

// Allows type H to be used with xml.Marshal
func (h xmlmap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestRenderYAML(t *testing.T) {
	w := httptest.NewRecorder()
	data := `
a : Easy!
b:
	c: 2
	d: [3, 4]
	`
	(YAML{data}).WriteContentType(w)
	assert.Equal(t, "application/yaml; charset=utf-8", w.Header().Get(literal_2953))

	err := (YAML{data}).Render(w)
	assert.NoError(t, err)
	assert.Equal(t, "|4-\n    a : Easy!\n    b:\n    \tc: 2\n    \td: [3, 4]\n    \t\n", w.Body.String())
	assert.Equal(t, "application/yaml; charset=utf-8", w.Header().Get(literal_2953))
}

type fail struct{}

// Hook MarshalYAML
func (ft *fail) MarshalYAML() (any, error) {
	return nil, errors.New("fail")
}

func TestRenderYAMLFail(t *testing.T) {
	w := httptest.NewRecorder()
	err := (YAML{&fail{}}).Render(w)
	assert.Error(t, err)
}

func TestRenderTOML(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}
	(TOML{data}).WriteContentType(w)
	assert.Equal(t, "application/toml; charset=utf-8", w.Header().Get(literal_2953))

	err := (TOML{data}).Render(w)
	assert.NoError(t, err)
	assert.Equal(t, "foo = 'bar'\nhtml = '<b>'\n", w.Body.String())
	assert.Equal(t, "application/toml; charset=utf-8", w.Header().Get(literal_2953))
}

func TestRenderTOMLFail(t *testing.T) {
	w := httptest.NewRecorder()
	err := (TOML{net.IPv4bcast}).Render(w)
	assert.Error(t, err)
}

// test Protobuf rendering
func TestRenderProtoBuf(t *testing.T) {
	w := httptest.NewRecorder()
	reps := []int64{int64(1), int64(2)}
	label := "test"
	data := &testdata.Test{
		Label: &label,
		Reps:  reps,
	}

	(ProtoBuf{data}).WriteContentType(w)
	protoData, err := proto.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "application/x-protobuf", w.Header().Get(literal_2953))

	err = (ProtoBuf{data}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, string(protoData), w.Body.String())
	assert.Equal(t, "application/x-protobuf", w.Header().Get(literal_2953))
}

func TestRenderProtoBufFail(t *testing.T) {
	w := httptest.NewRecorder()
	data := &testdata.Test{}
	err := (ProtoBuf{data}).Render(w)
	assert.Error(t, err)
}

func TestRenderXML(t *testing.T) {
	w := httptest.NewRecorder()
	data := xmlmap{
		"foo": "bar",
	}

	(XML{data}).WriteContentType(w)
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get(literal_2953))

	err := (XML{data}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "<map><foo>bar</foo></map>", w.Body.String())
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get(literal_2953))
}

func TestRenderRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "/test-redirect", nil)
	assert.NoError(t, err)

	data1 := Redirect{
		Code:     http.StatusMovedPermanently,
		Request:  req,
		Location: literal_3602,
	}

	w := httptest.NewRecorder()
	err = data1.Render(w)
	assert.NoError(t, err)

	data2 := Redirect{
		Code:     http.StatusOK,
		Request:  req,
		Location: literal_3602,
	}

	w = httptest.NewRecorder()
	assert.PanicsWithValue(t, "Cannot redirect with status code 200", func() {
		err := data2.Render(w)
		assert.NoError(t, err)
	})

	data3 := Redirect{
		Code:     http.StatusCreated,
		Request:  req,
		Location: literal_3602,
	}

	w = httptest.NewRecorder()
	err = data3.Render(w)
	assert.NoError(t, err)

	// only improve coverage
	data2.WriteContentType(w)
}

func TestRenderData(t *testing.T) {
	w := httptest.NewRecorder()
	data := []byte(literal_5480)

	err := (Data{
		ContentType: literal_8670,
		Data:        data,
	}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, literal_5480, w.Body.String())
	assert.Equal(t, literal_8670, w.Header().Get(literal_2953))
}

func TestRenderString(t *testing.T) {
	w := httptest.NewRecorder()

	(String{
		Format: "hello %s %d",
		Data:   []any{},
	}).WriteContentType(w)
	assert.Equal(t, literal_6740, w.Header().Get(literal_2953))

	err := (String{
		Format: literal_2586,
		Data:   []any{"manu", 2},
	}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "hola manu 2", w.Body.String())
	assert.Equal(t, literal_6740, w.Header().Get(literal_2953))
}

func TestRenderStringLenZero(t *testing.T) {
	w := httptest.NewRecorder()

	err := (String{
		Format: literal_2586,
		Data:   []any{},
	}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, literal_2586, w.Body.String())
	assert.Equal(t, literal_6740, w.Header().Get(literal_2953))
}

func TestRenderHTMLTemplate(t *testing.T) {
	w := httptest.NewRecorder()
	templ := template.Must(template.New("t").Parse(`Hello {{.name}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("t", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "Hello alexandernyquist", w.Body.String())
	assert.Equal(t, literal_1906, w.Header().Get(literal_2953))
}

func TestRenderHTMLTemplateEmptyName(t *testing.T) {
	w := httptest.NewRecorder()
	templ := template.Must(template.New("").Parse(`Hello {{.name}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "Hello alexandernyquist", w.Body.String())
	assert.Equal(t, literal_1906, w.Header().Get(literal_2953))
}

func TestRenderHTMLDebugFiles(t *testing.T) {
	w := httptest.NewRecorder()
	htmlRender := HTMLDebug{
		Files:   []string{"../testdata/template/hello.tmpl"},
		Glob:    "",
		Delims:  Delims{Left: "{[{", Right: "}]}"},
		FuncMap: nil,
	}
	instance := htmlRender.Instance("hello.tmpl", map[string]any{
		"name": "thinkerou",
	})

	err := instance.Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "<h1>Hello thinkerou</h1>", w.Body.String())
	assert.Equal(t, literal_1906, w.Header().Get(literal_2953))
}

func TestRenderHTMLDebugGlob(t *testing.T) {
	w := httptest.NewRecorder()
	htmlRender := HTMLDebug{
		Files:   nil,
		Glob:    "../testdata/template/hello*",
		Delims:  Delims{Left: "{[{", Right: "}]}"},
		FuncMap: nil,
	}
	instance := htmlRender.Instance("hello.tmpl", map[string]any{
		"name": "thinkerou",
	})

	err := instance.Render(w)

	assert.NoError(t, err)
	assert.Equal(t, "<h1>Hello thinkerou</h1>", w.Body.String())
	assert.Equal(t, literal_1906, w.Header().Get(literal_2953))
}

func TestRenderHTMLDebugPanics(t *testing.T) {
	htmlRender := HTMLDebug{
		Files:   nil,
		Glob:    "",
		Delims:  Delims{"{{", "}}"},
		FuncMap: nil,
	}
	assert.Panics(t, func() { htmlRender.Instance("", nil) })
}

func TestRenderReader(t *testing.T) {
	w := httptest.NewRecorder()

	body := literal_5480
	headers := make(map[string]string)
	headers[literal_8690] = `attachment; filename="filename.png"`
	headers[literal_7834] = "requestId"

	err := (Reader{
		ContentLength: int64(len(body)),
		ContentType:   literal_8670,
		Reader:        strings.NewReader(body),
		Headers:       headers,
	}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, body, w.Body.String())
	assert.Equal(t, literal_8670, w.Header().Get(literal_2953))
	assert.Equal(t, strconv.Itoa(len(body)), w.Header().Get("Content-Length"))
	assert.Equal(t, headers[literal_8690], w.Header().Get(literal_8690))
	assert.Equal(t, headers[literal_7834], w.Header().Get(literal_7834))
}

func TestRenderReaderNoContentLength(t *testing.T) {
	w := httptest.NewRecorder()

	body := literal_5480
	headers := make(map[string]string)
	headers[literal_8690] = `attachment; filename="filename.png"`
	headers[literal_7834] = "requestId"

	err := (Reader{
		ContentLength: -1,
		ContentType:   literal_8670,
		Reader:        strings.NewReader(body),
		Headers:       headers,
	}).Render(w)

	assert.NoError(t, err)
	assert.Equal(t, body, w.Body.String())
	assert.Equal(t, literal_8670, w.Header().Get(literal_2953))
	assert.NotContains(t, "Content-Length", w.Header())
	assert.Equal(t, headers[literal_8690], w.Header().Get(literal_8690))
	assert.Equal(t, headers[literal_7834], w.Header().Get(literal_7834))
}

func TestRenderWriteError(t *testing.T) {
	data := []interface{}{"value1", "value2"}
	prefix := "my-prefix:"
	r := SecureJSON{Data: data, Prefix: prefix}
	ew := &errorWriter{
		bufString:        prefix,
		ResponseRecorder: httptest.NewRecorder(),
	}
	err := r.Render(ew)
	assert.NotNil(t, err)
	assert.Equal(t, `write "my-prefix:" error`, err.Error())
}

const literal_3516 = "application/json; charset=utf-8"

const literal_2953 = "Content-Type"

const literal_9015 = "while(1);"

const literal_8096 = "application/javascript; charset=utf-8"

const literal_3602 = "/new/location"

const literal_5480 = "#!PNG some raw data"

const literal_8670 = "image/png"

const literal_6740 = "text/plain; charset=utf-8"

const literal_2586 = "hola %s %d"

const literal_1906 = "text/html; charset=utf-8"

const literal_8690 = "Content-Disposition"

const literal_7834 = "x-request-id"
