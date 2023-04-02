package client

import (
	"bytes"
	"io"
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
	rw.log.Info(rw.buf.String())
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.RoundTripper.RoundTrip(req)
	rt.LogRequest(req)
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
		rt.log.Error(err.Error())
	}
	return resp, err
}

func (rt *roundTripper) LogRequest(req *http.Request) {
	rt.log.Info("%s %s", req.Method, req.URL)
}

func (rt *roundTripper) LogResponse(resp *http.Response) {
	if resp == nil {
		return
	}
	rt.log.Info("%s %s", resp.Proto, resp.Status)
}
