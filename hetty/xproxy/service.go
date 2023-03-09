package xproxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kataras/golog"
	"net/http"
	"net/http/httputil"
)

type Serv struct {
	serv *http.Server
}

func NewServ(addr string) (*Serv, error) {
	// Load existing CA certificate and key from disk, or generate and write to disk if no files exist yet.
	caCert, caKey, err := LoadOrCreateCA(caKeyPath, caCertPath)
	if err != nil {
		return nil, err
	}
	proxy, err := NewProxy(Config{
		CACert: caCert,
		CAKey:  caKey,
		Logger: golog.Default,
	})
	if err != nil {
		return nil, err
	}
	proxy.UseRequestModifier(RequestModifier)
	proxy.UseResponseModifier(ResponseModifier)

	router := mux.NewRouter().SkipClean(true)
	// Fallback (default) is the Proxy handler.
	router.PathPrefix("").Handler(proxy)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
	}
	return &Serv{serv: httpServer}, nil
}

func (p *Serv) Listen(ctx context.Context) {

	go func() {
		golog.Infof("proxy listen at: %s", p.serv.Addr)
		err := p.serv.ListenAndServe()
		if err != http.ErrServerClosed {
			golog.Fatal("proxy listen server closed unexpected.")
		}
	}()
}

func (p *Serv) Stop() error {
	return p.serv.Shutdown(context.Background())
}

func RequestModifier(next RequestModifyFunc) RequestModifyFunc {
	return func(req *http.Request) {
		modifiedReq, err := interceptRequest(req.Context(), req)
		if err != nil {
			golog.Errorf("Failed to intercept request.", "error", err)
		}
		*req = *modifiedReq
		next(req)
	}
}

func interceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	//reqID, ok := RequestIDFromContext(ctx)
	//if !ok {
	//	return req, errors.New("failed to intercept: context doesn't have an ID")
	//}
	// todo add filter
	//golog.Infof("id:%v host:%s uri:%v", reqID, req.Host, req.URL.String())
	return req, nil
}

func ResponseModifier(next ResponseModifyFunc) ResponseModifyFunc {
	return func(res *http.Response) error {
		// This is a blocking operation, that gets unblocked when either a modified response is returned or an error.
		//nolint:bodyclose
		modifiedRes, err := interceptResponse(res.Request.Context(), res)
		if err != nil {
			return fmt.Errorf("failed to intercept response: %w", err)
		}

		*res = *modifiedRes

		return next(res)
	}
}

// InterceptResponse adds an HTTP response to an array of pending intercepted responses, alongside channels used for
// sending a cancellation signal and receiving a modified response. It's safe for concurrent use.
func interceptResponse(ctx context.Context, res *http.Response) (*http.Response, error) {
	//reqID, ok := RequestIDFromContext(ctx)
	//if !ok {
	//	return res, errors.New("failed to intercept: context doesn't have an ID")
	//}
	// todo add filter
	dumpRequest, err := httputil.DumpRequest(res.Request, res.Request.Body != nil)
	if err != nil {
		golog.Error(err)
	}
	dumpResponse, err := httputil.DumpResponse(res, res.Body != nil)
	if err != nil {
		golog.Error(err)
	}
	golog.Infof("\n——————————————————————————————————\n%s\n%s\n", string(dumpRequest), string(dumpResponse))

	return res, nil
}
