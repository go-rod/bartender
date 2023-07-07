package bartender_test

import (
	"net/http"
	"testing"

	"github.com/go-rod/bartender"
	"github.com/ysmood/got"
)

func TestBasic(t *testing.T) {
	g := got.T(t)

	website := g.Serve().Route("/", ".html", `<html>
		<body></body>
		<script>
			window.onload = () => {
				document.body.innerHTML = location.pathname + location.search
			}
		</script>
	</html>`)

	proxy := g.Serve()

	bt := bartender.New("", website.URL())

	proxy.Mux.HandleFunc("/", bt.ServeHTTP)

	{
		// browser
		res := g.Req("", proxy.URL("/test?q=ok"), http.Header{"Accept-Language": {"en"}})
		g.Has(res.String(), "<body></body>")
	}

	{
		// web crawler
		res := g.Req("", proxy.URL("/test?q=ok"))
		g.Has(res.String(), "/test?q=ok")
	}
}
