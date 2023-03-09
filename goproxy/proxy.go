package gproxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/elazarl/goproxy"
	"github.com/kataras/golog"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
)

type Proxy struct {
	addr string
	serv *http.Server
}

func NewProxy(addr string) *Proxy {
	handler := newGoProxyHandler()
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
	}
	return &Proxy{serv: httpServer, addr: addr}
}

func (p *Proxy) Listen(ctx context.Context) {
	go func() {
		golog.Infof("proxy listen at: %s", p.addr)
		err := p.serv.ListenAndServe()
		if err != http.ErrServerClosed {
			golog.Warnf("proxy listen server closed unexpected.")
		}
	}()
}

func (p *Proxy) Stop() error {
	golog.Infof("proxy already closed: %s", p.addr)
	return p.serv.Shutdown(context.Background())
}

func newGoProxyHandler() *goproxy.ProxyHttpServer {

	proxy := goproxy.NewProxyHttpServer()
	// debug 模式
	//proxy.Verbose = true

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	//request filter
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// transform body
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				golog.Errorf("read req body failed, ", err)
				return nil, nil
			}
			bodyStr := string(body)
			req.Body = ReaderCloser(bytes.NewBuffer([]byte(bodyStr)))
			return req, nil
		})
	// response filter
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if ctx.Resp.Body != nil {
			ctx.Resp.Body = ReaderCloser(ctx.Resp.Body)
		}
		dumpRequest, err := httputil.DumpRequest(ctx.Req, ctx.Req.Body != nil)
		if err != nil {
			golog.Error(err)
		}
		dumpResponse, err := httputil.DumpResponse(ctx.Resp, ctx.Resp.Body != nil)
		if err != nil {
			golog.Error(err)
		}
		golog.Infof("\n——————————————————————————————————\n%s\n%s\n", string(dumpRequest), string(dumpResponse))

		return resp
	})

	return proxy
}

func setCA() error {

	pwd, _ := os.Getwd()
	caCert, _ := ioutil.ReadFile(pwd + "/ca.pem")
	caKey, _ := ioutil.ReadFile(pwd + "/ca.key.pem")

	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
