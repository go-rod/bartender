// Package bartender is a service to make web crawlers consume webpages easier
package bartender

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-rod/rod"
	"github.com/mileusna/useragent"
)

type Bartender struct {
	addr       string
	target     *url.URL
	proxy      *httputil.ReverseProxy
	bypassList map[string]bool
	pool       rod.PagePool
}

func New(addr, target string, poolSize int) *Bartender {
	u, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &Bartender{
		addr:   addr,
		target: u,
		proxy:  proxy,
		bypassList: map[string]bool{
			useragent.Opera:            true,
			useragent.OperaMini:        true,
			useragent.OperaTouch:       true,
			useragent.Chrome:           true,
			useragent.HeadlessChrome:   true,
			useragent.Firefox:          true,
			useragent.InternetExplorer: true,
			useragent.Safari:           true,
			useragent.Edge:             true,
			useragent.Vivaldi:          true,
		},
		pool: rod.NewPagePool(poolSize),
	}
}

func (b *Bartender) BypassUserAgentNames(list map[string]bool) {
	b.bypassList = list
}

func (b *Bartender) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ua := useragent.Parse(r.Header.Get("User-Agent"))
	if r.Method != http.MethodGet || b.bypassList[ua.Name] {
		b.proxy.ServeHTTP(w, r)

		return
	}

	if b.RenderPage(w, r) {
		return
	}

	b.proxy.ServeHTTP(w, r)
}

// RenderPage returns true if the page is rendered by the headless browser.
func (b *Bartender) RenderPage(w http.ResponseWriter, r *http.Request) bool {
	u := b.getTargetURL(r.URL)

	statusCode, resHeader := getHeader(r.Context(), u)

	if !htmlContentType(resHeader) {
		return false
	}

	log.Println("headless render:", u)

	page := b.pool.Get(func() *rod.Page { return rod.New().MustConnect().MustPage() })
	defer b.pool.Put(page)

	page.MustNavigate(u).MustWaitStable()

	for k, vs := range resHeader {
		if k == "Content-Length" {
			continue
		}

		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(statusCode)

	_, err := w.Write([]byte(page.MustHTML()))
	if err != nil {
		panic(err)
	}

	return true
}

func (b *Bartender) getTargetURL(reqURL *url.URL) string {
	u := *reqURL
	u.Scheme = b.target.Scheme
	u.Host = b.target.Host

	return u.String()
}

func getHeader(ctx context.Context, u string) (int, http.Header) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	_ = res.Body.Close()

	return res.StatusCode, res.Header
}

func htmlContentType(h http.Header) bool {
	return strings.Contains(h.Get("Content-Type"), "text/html")
}
