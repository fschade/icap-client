package icapclient

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestSetEncapsulatedHeaderValue(t *testing.T) {
	type testSample struct {
		icapReqStr  string
		httpReqStr  string
		httpRespStr string
		result      string
	}

	sampleTable := []testSample{
		{
			icapReqStr: "REQMOD\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "GET / HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Cookie: ff39fk3jur@4ii0e02i\r\n" +
				"If-None-Match: \"xyzzy\", \"r2d2xxxx\"\r\n\r\n",
			httpRespStr: "",
			result:      "REQMOD\r\nEncapsulated:  req-hdr=0, null-body=170\r\n\r\n",
		},
		{
			icapReqStr: "REQMOD\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Pragma: no-cache\r\n\r\n" +
				"1e\r\n" +
				"I am posting this information.\r\n" +
				"0\r\n\r\n",
			result: "REQMOD\r\nEncapsulated:  req-hdr=0, req-body=147\r\n\r\n",
		},
		{
			icapReqStr: "RESPMOD\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "GET /origin-resource HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain, image/gif\r\n" +
				"Accept-Encoding: gzip, compress\r\n\r\n",
			httpRespStr: "HTTP/1.1 200 OK\r\n" +
				"Date: Mon, 10 Jan 2000 09:52:22 GMT\r\n" +
				"Server: Apache/1.3.6 (Unix)\r\n" +
				"ETag: \"63840-1ab7-378d415b\"\r\n" +
				"Content-Type: text/html\r\n" +
				"Content-Length: 51\r\n\r\n" +
				"33\r\n" +
				"This is data that was returned by an origin server.\r\n" +
				"0\r\n\r\n",
			result: "RESPMOD\r\nEncapsulated:  req-hdr=0, res-hdr=137, res-body=296\r\n\r\n",
		},
		{
			icapReqStr: "RESPMOD\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Pragma: no-cache\r\n\r\n" +
				"1e\r\n" +
				"I am posting this information.\r\n" +
				"0\r\n\r\n",
			httpRespStr: "HTTP/1.1 200 OK\r\n" +
				"Date: Mon, 10 Jan 2000 09:52:22 GMT\r\n" +
				"Server: Apache/1.3.6 (Unix)\r\n" +
				"ETag: \"63840-1ab7-378d415b\"\r\n" +
				"Content-Type: text/html\r\n" +
				"Content-Length: 51\r\n\r\n" +
				"33\r\n" +
				"This is data that was returned by an origin server.\r\n" +
				"0\r\n\r\n",
			result: "RESPMOD\r\nEncapsulated:  req-hdr=0, req-body=147, res-hdr=188, res-body=347\r\n\r\n",
		},
		{
			icapReqStr: "RESPMOD\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Pragma: no-cache\r\n\r\n" +
				"1e\r\n" +
				"I am posting this information.\r\n" +
				"0\r\n\r\n",
			httpRespStr: "HTTP/1.1 200 OK\r\n" +
				"Date: Mon, 10 Jan 2000 09:52:22 GMT\r\n" +
				"Server: Apache/1.3.6 (Unix)\r\n" +
				"ETag: \"63840-1ab7-378d415b\"\r\n" +
				"Content-Type: text/html\r\n" +
				"Content-Length: 51\r\n\r\n",
			result: "RESPMOD\r\nEncapsulated:  req-hdr=0, req-body=147, res-hdr=188, null-body=347\r\n\r\n",
		},
		{
			icapReqStr:  "OPTIONS\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr:  "",
			httpRespStr: "",
			result:      "OPTIONS\r\nEncapsulated:  null-body=0\r\n\r\n",
		},
		{
			icapReqStr: "OPTIONS\r\nEncapsulated: %s\r\n\r\n",
			httpReqStr: "GET /origin-resource HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain, image/gif\r\n" +
				"Accept-Encoding: gzip, compress\r\n\r\n",
			httpRespStr: "",
			result:      "OPTIONS\r\nEncapsulated:  opt-body=0\r\n\r\n",
		},
	}

	for _, sample := range sampleTable {
		setEncapsulatedHeaderValue(&sample.icapReqStr, sample.httpReqStr, sample.httpRespStr)
		if sample.icapReqStr != sample.result {
			t.Logf("Wanted icap message after setting encapsulation: %s , got:%s", sample.result, sample.icapReqStr)
			t.Fail()
		}
	}
}

func TestAddHexaBodyByteNotations(t *testing.T) {
	type testSample struct {
		msg    string
		result string
	}

	sampleTable := []testSample{
		{
			msg:    "Hello World!",
			result: "c\r\nHello World!\r\n0\r\n",
		},
		{
			msg:    "This is another message. Alright bye!",
			result: "25\r\nThis is another message. Alright bye!\r\n0\r\n",
		},
	}

	for _, sample := range sampleTable {
		addHexBodyByteNotations(&sample.msg)
		if sample.msg != sample.result {
			t.Logf("Wanted message after adding hexa body notations: %s, got:%s", sample.result, sample.msg)
			t.Fail()
		}
	}
}

func TestParsePreviewBodyBytes(t *testing.T) {
	type testSample struct {
		previewBytes int
		httpMsg      string
		result       string
	}

	sampleTable := []testSample{
		{
			previewBytes: 10,
			httpMsg: "HTTP/1.1 200 OK\r\n" +
				"Date: Mon, 10 Jan 2000 09:52:22 GMT\r\n" +
				"Server: Apache/1.3.6 (Unix)\r\n" +
				"ETag: \"63840-1ab7-378d415b\"\r\n" +
				"Content-Type: text/html\r\n" +
				"Content-Length: 51\r\n\r\n" +
				"This is data that was returned by an origin server.\r\n\r\n",
			result: "HTTP/1.1 200 OK\r\n" +
				"Date: Mon, 10 Jan 2000 09:52:22 GMT\r\n" +
				"Server: Apache/1.3.6 (Unix)\r\n" +
				"ETag: \"63840-1ab7-378d415b\"\r\n" +
				"Content-Type: text/html\r\n" +
				"Content-Length: 51\r\n\r\n" +
				"This is da",
		},
		{
			previewBytes: 10,
			httpMsg: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Pragma: no-cache\r\n\r\n" +
				"I am posting this information.\r\n",
			result: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
				"Host: www.origin-server.com\r\n" +
				"Accept: text/html, text/plain\r\n" +
				"Accept-Encoding: compress\r\n" +
				"Pragma: no-cache\r\n\r\n" +
				"I am posti",
		},
	}

	for _, sample := range sampleTable {
		parsePreviewBodyBytes(&sample.httpMsg, sample.previewBytes)
		if sample.httpMsg != sample.result {
			t.Logf("Wanted http message after parsing to be: %s , got: %s", sample.result, sample.httpMsg)
			t.Fail()
		}
	}
}

func TestToICAPMessage(t *testing.T) {
	t.Run("MethodOPTIONS", func(t *testing.T) {

		req, _ := NewRequest(context.Background(), MethodOPTIONS, "icap://localhost:1344/something", nil, nil)

		message, err := toICAPMessage(req)

		if err != nil {
			t.Fatal(err.Error())
		}

		wanted := "OPTIONS icap://localhost:1344/something ICAP/1.0\r\n" +
			"Encapsulated:  null-body=0\r\n\r\n"

		got := string(message)

		if wanted != got {
			t.Logf("wanted: %s, got: %s\n", wanted, got)
			t.Fail()
		}

	})

	t.Run("MethodREQMOD", func(t *testing.T) { // FIXME: add proper wanted string and complete this unit test
		httpReq, _ := http.NewRequest(http.MethodGet, "http://someurl.com", nil)

		req, _ := NewRequest(context.Background(), MethodREQMOD, "icap://localhost:1344/something", httpReq, nil)

		message, err := toICAPMessage(req)
		if err != nil {
			t.Fatal(err.Error())
		}

		wanted := "REQMOD icap://localhost:1344/something ICAP/1.0\r\n" +
			"Encapsulated:  req-hdr=0, null-body=109\r\n\r\n" +
			"GET http://someurl.com HTTP/1.1\r\n" +
			"Host: someurl.com\r\n" +
			"User-Agent: Go-http-client/1.1\r\n" +
			"Accept-Encoding: gzip\r\n\r\n"

		got := string(message)

		if wanted != got {
			t.Logf("wanted: \n%s\ngot: \n%s\n", wanted, got)
			t.Fail()
		}

		httpReq, _ = http.NewRequest(http.MethodPost, "http://someurl.com", bytes.NewBufferString("Hello World"))

		req, _ = NewRequest(context.Background(), MethodREQMOD, "icap://localhost:1344/something", httpReq, nil)

		message, err = toICAPMessage(req)
		if err != nil {
			t.Fatal(err.Error())
		}

		wanted = "REQMOD icap://localhost:1344/something ICAP/1.0\r\n" +
			"Encapsulated:  req-hdr=0, req-body=130\r\n\r\n" +
			"POST http://someurl.com HTTP/1.1\r\n" +
			"Host: someurl.com\r\n" +
			"User-Agent: Go-http-client/1.1\r\n" +
			"Content-Length: 11\r\n" +
			"Accept-Encoding: gzip\r\n\r\n" +
			"b\r\n" +
			"Hello World\r\n" +
			"0\r\n\r\n"

		got = string(message)

		if wanted != got {
			t.Logf("wanted: \n%s\ngot: \n%s\n", wanted, got)
			t.Fail()
		}
	})

	t.Run("MethodRESPMOD", func(t *testing.T) {
		httpReq, _ := http.NewRequest(http.MethodPost, "http://someurl.com", bytes.NewBufferString("Hello World"))
		httpResp := &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Header: http.Header{
				"Content-Type":   []string{"plain/text"},
				"Content-Length": []string{"11"},
			},
			ContentLength: 11,
			Body:          ioutil.NopCloser(strings.NewReader("Hello World")),
		}

		req, _ := NewRequest(context.Background(), MethodRESPMOD, "icap://localhost:1344/something", httpReq, httpResp)

		message, err := toICAPMessage(req)
		if err != nil {
			t.Fatal(err.Error())
		}

		wanted := "RESPMOD icap://localhost:1344/something ICAP/1.0\r\n" +
			"Encapsulated:  req-hdr=0, req-body=130, res-hdr=145, res-body=210\r\n\r\n" +
			"POST http://someurl.com HTTP/1.1\r\n" +
			"Host: someurl.com\r\n" +
			"User-Agent: Go-http-client/1.1\r\n" +
			"Content-Length: 11\r\n" +
			"Accept-Encoding: gzip\r\n\r\n" +
			"Hello World\r\n\r\n" +
			"HTTP/1.0 200 OK\r\n" +
			"Content-Length: 11\r\n" +
			"Content-Type: plain/text\r\n\r\n" +
			"b\r\n" +
			"Hello World\r\n" +
			"0\r\n\r\n"

		got := string(message)

		if wanted != got {
			t.Logf("wanted: \n%s\ngot: \n%s\n", wanted, got)
			t.Fail()
		}
	})
}
