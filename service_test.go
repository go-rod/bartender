package bartender_test

import (
	"net/http"
	"testing"

	"github.com/go-rod/bartender"
	"github.com/ysmood/got"
)

func TestBasic(t *testing.T) {
	g := got.T(t)

	website := g.Serve()

	website.Route("/a.png", ".png", "image")
	website.Route("/", ".html", `<html>
		<body></body>
		<script>
			window.onload = () => {
				document.body.innerHTML = location.pathname + location.search
			}
		</script>
	</html>`)

	proxy := g.Serve()

	bt := bartender.New("", website.URL(), 2)

	proxy.Mux.HandleFunc("/", bt.ServeHTTP)

	{
		//nolint: lll
		// browser
		res := g.Req("", proxy.URL("/test?q=ok"), http.Header{"User-Agent": {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"}})
		g.Has(res.String(), "<body></body>")
	}

	{
		// web crawler
		res := g.Req("", proxy.URL("/test?q=ok"))
		g.Has(res.String(), "/test?q=ok")
	}

	{
		// can get image
		res := g.Req("", proxy.URL("/a.png"))
		g.Has(res.String(), "image")
	}
}
