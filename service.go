// Package bartender is a service to make web crawlers consume webpages easier
package bartender

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/mileusna/useragent"
)

type Bartender struct {
	addr       string
	target     *url.URL
	proxy      *httputil.ReverseProxy
	bypassList map[string]bool
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
	}
}

func (b *Bartender) BypassUserAgentNames(list map[string]bool) {
	b.bypassList = list
}

func (b *Bartender) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ua := useragent.Parse(r.Header.Get("User-Agent"))
	if b.bypassList[ua.Name] {
		b.proxy.ServeHTTP(w, r)

		return
	}

	b.RenderPage(w, r)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, _ = w.Write([]byte(page.MustHTML()))
}
