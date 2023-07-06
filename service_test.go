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

	proxy.Mux.HandleFunc("/", bt.Handler)

	{
		res := g.Req("", proxy.URL("/test?q=ok"))
		g.Has(res.String(), "<body></body>")
	}

	{
		res := g.Req("", proxy.URL("/test?q=ok"), http.Header{
			"User-Agent": []string{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/113.0.5672.127 Safari/537.36"},
		})
		g.Has(res.String(), "/test?q=ok")
	}
}
