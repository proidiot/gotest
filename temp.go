package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type crawlee struct {
	depth int;
	url string;
	children []string;
}

func deepen(pile map[string]crawlee, url string, foo func(string)) {
	cr := pile[url];
	cr.depth += 1;
	if cr.depth == 1 {
		foo(url);
	} else {
		for _, u := range cr.children {
			deepen(pile, u, foo);
		}
	}
}

func crawl(url string, depth int, fetcher Fetcher, ch chan crawlee) {
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		ch <- crawlee{depth, url, nil};
	} else {
		fmt.Printf("found: %s %q\n", url, body);
		ch <- crawlee{depth, url, urls};
	}
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	ch := make(chan crawlee);
	pile := make(map[string]crawlee);
	unfin := make(map[string]bool);

	unfin[url] = true;
	go crawl(url, depth, fetcher, ch);

	for cr := range ch {
		pile[cr.url] = cr;

		delete(unfin, cr.url);

		for _, u := range cr.children {
			if unfin[u] == false {
				tcs := pile[u];
				if tcs.depth == 0 {
					unfin[u] = true;
					go crawl(u, cr.depth - 1, fetcher, ch);
				} else {
					for tcs.depth < cr.depth - 1 {
						foo := func(tu string) {
							unfin[tu] = true;
							go crawl(tu, 1, fetcher, ch);
						}
						deepen(pile, u, foo);
						tcs = pile[u];
					}
				}
			}
		}

		if len(unfin) == 0 {
			close(ch);
		}
	}
}

func main() {
	Crawl("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

