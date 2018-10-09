package main

import (
	"log"
	"fmt"
	"net/http"
	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/html"
)

const (
	// Service - use a tracer level global service name
	Service = "FindLinksFork"
)

func main() {
	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  Service,
		LogLevel: instana.Debug}))
	http.HandleFunc("/", handler)
	
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

func handler(w http.ResponseWriter, req *http.Request) {
	
	wireContext, _ := ot.GlobalTracer().Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	parentSpan := ot.GlobalTracer().StartSpan("server", ext.RPCServerOption(wireContext))
	parentSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	parentSpan.SetTag(string(ext.PeerHostname), req.Host)
	parentSpan.SetTag(string(ext.HTTPUrl), req.URL.Path)
	parentSpan.SetTag(string(ext.HTTPMethod), req.Method)
	parentSpan.SetTag(string(ext.HTTPStatusCode), 200)
	
	
	url := req.URL.Query().Get("q")
	fmt.Fprintf(w, "Page = %q\n", url)
	if len(url) == 0 {
		return
	}
	page, err := parse("https://" + url)
	if err != nil {
		fmt.Printf("Error getting page %s %s\n", url, err)
		return
	}
	links := pageLinks(nil, page)
	for _, link := range links {
		fmt.Fprintf(w, "Link = %q\n", link)
	}
	parentSpan.Finish()
}

func parse(url string) (*html.Node, error) {
	fmt.Println(url)
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Cannot get page")
	}
	b, err := html.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse page")
	}
	return b, err
}

func pageLinks(links []string, n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				links = append(links, a.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = pageLinks(links, c)
	}
	return links
}
