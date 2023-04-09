package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/collinvandyck/gpterm/lib/log"
)

type roundTripper struct {
	http.RoundTripper
	log log.Logger
}

type recordingWriter struct {
	buf bytes.Buffer
	log log.Logger
	io.Writer
}

func (rw *recordingWriter) Write(p []byte) (int, error) {
	rw.buf.Write(p)
	return rw.Writer.Write(p)
}

func (rw *recordingWriter) Sync() {
	rw.log.Log("Response body", "data", rw.buf.String())
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.LogRequest(req)
	resp, err := rt.RoundTripper.RoundTrip(req)
	rt.LogResponse(resp)
	if resp != nil {
		recorder := &recordingWriter{log: rt.log}
		var body io.ReadCloser = resp.Body
		pr, pw := io.Pipe()
		recorder.Writer = pw
		resp.Body = pr
		go func() {
			io.Copy(recorder, body)
			pw.Close()
			recorder.Sync()
		}()
	}
	if err != nil {
		rt.log.Log("Request failed", "err", err)
	}
	return resp, err
}

func (rt *roundTripper) LogRequest(req *http.Request) {
	bs, err := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(bs))
	rt.log.Log("Request", "method", req.Method, "url", req.URL, "body", string(bs), "err", err)
}

func (rt *roundTripper) LogResponse(resp *http.Response) {
	if resp == nil {
		return
	}
	rt.log.Log("Response", "status", resp.Status, "proto", resp.Proto)
}
