// Package bartender is a service to make web crawlers consume webpages easier
package bartender

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/mileusna/useragent"
)

var DefaultBypassUserAgentNames = []string{
	useragent.Opera,
	useragent.OperaMini,
	useragent.OperaTouch,
	useragent.Chrome,
	useragent.HeadlessChrome,
	useragent.Firefox,
	useragent.InternetExplorer,
	useragent.Safari,
	useragent.Edge,
	useragent.Vivaldi,
}

type Bartender struct {
	addr          string
	target        *url.URL
	proxy         *httputil.ReverseProxy
	bypassList    map[string]bool
	pool          rod.PagePool
	blockRequests []string
	maxWait       time.Duration
}

func New(addr, target string, poolSize int) *Bartender {
	u, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &Bartender{
		addr:          addr,
		target:        u,
		proxy:         proxy,
		bypassList:    strToMap(DefaultBypassUserAgentNames),
		pool:          rod.NewPagePool(poolSize),
		blockRequests: []string{},
		maxWait:       3 * time.Second,
	}
}

func (b *Bartender) BypassUserAgentNames(list ...string) {
	b.bypassList = strToMap(list)
}

func (b *Bartender) BlockRequests(patterns ...string) {
	b.blockRequests = patterns
}

// MaxWait sets the max wait time for the headless browser to render the page.
// If the max wait time is reached, bartender will stop waiting for page rendering and
// immediately return the current html.
func (b *Bartender) MaxWait(d time.Duration) {
	b.maxWait = d
}

func (b *Bartender) getPage() *rod.Page {
	return b.pool.Get(b.newPage)
}

func (b *Bartender) newPage() *rod.Page {
	page := rod.New().MustConnect().MustPage()

	if len(b.blockRequests) > 0 {
		router := page.HijackRequests()

		for _, pattern := range b.blockRequests {
			router.MustAdd(pattern, func(ctx *rod.Hijack) {
				ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			})
		}

		go router.Run()
	}

	log.Println("headless browser started:", page.SessionID)

	return page
}

// WarmUp pre-creates the headless browsers.
func (b *Bartender) WarmUp() {
	for i := 0; i < len(b.pool); i++ {
		b.pool.Put(b.getPage())
	}
}

// AutoFree automatically closes the each headless browser after a period of time.
// It prevent the memory leak of the headless browser.
func (b *Bartender) AutoFree() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)

			err := b.getPage().Browser().Close()
			if err != nil {
				log.Println("failed to close browser:", err)

				continue
			}
			b.pool.Put(nil)
		}
	}()
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

	for k, vs := range resHeader {
		if k == "Content-Length" {
			continue
		}

		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(statusCode)

	page := b.getPage()
	defer b.pool.Put(page)

	page, cancel := page.Context(r.Context()).WithCancel()

	once := sync.Once{}

	go func() {
		time.Sleep(b.maxWait)
		once.Do(func() {
			log.Println("max wait time reached, return current html:", u)
			body, _ := page.HTML()
			_, _ = w.Write([]byte(body))
			cancel()
		})
	}()

	_ = page.Navigate(u)

	_ = page.WaitStable(time.Second)

	body, _ := page.HTML()

	once.Do(func() {
		log.Println("headless render done:", u)
		_, _ = w.Write([]byte(body))
	})

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

func strToMap(list []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range list {
		m[s] = true
	}

	return m
}
