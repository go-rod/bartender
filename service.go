// Package bartender is a service to make web crawlers consume webpages easier
package bartender

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Bartender struct {
	addr   string
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func New(addr, target string) *Bartender {
	u, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &Bartender{
		addr:   addr,
		target: u,
		proxy:  proxy,
	}
}

func (b *Bartender) Serve() {
	http.HandleFunc("/", b.Handler)

	err := http.ListenAndServe(b.addr, nil)
	if err != nil {
		panic(err)
	}
}

func (b *Bartender) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.Header.Get("Accept-Language") == "" {
		b.RenderPage(w, r)

		return
	}

	b.proxy.ServeHTTP(w, r)
}

func (b *Bartender) RenderPage(w http.ResponseWriter, r *http.Request) {
	log.Println("headless render:", r.URL.String())

	u := *r.URL
	u.Scheme = b.target.Scheme
	u.Host = b.target.Host

	l := launcher.New()
	defer l.Cleanup()

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(u.String()).MustWaitStable()

	_, _ = w.Write([]byte(page.MustHTML()))
}
