package bartender

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/mileusna/useragent"
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
	http.HandleFunc("/", b.handler)

	err := http.ListenAndServe(b.addr, nil)
	if err != nil {
		panic(err)
	}
}

func (b *Bartender) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ua := useragent.Parse(r.UserAgent())

		if ua.Bot {
			b.renderPage(w, r)
			return
		}
	}

	b.proxy.ServeHTTP(w, r)
}

func (b *Bartender) renderPage(w http.ResponseWriter, r *http.Request) {
	u := *r.URL

	u.Scheme = b.target.Scheme
	u.Host = b.target.Host

	res, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}
	_ = res.Body.Close()

	w.WriteHeader(res.StatusCode)

	l := launcher.New()
	defer l.Cleanup()

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(u.String()).MustWaitStable()

	_, _ = w.Write([]byte(page.MustHTML()))
}
