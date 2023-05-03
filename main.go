package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"os"
	"regexp"
	"strings"
)

var articleLinks []string
var links []string
var topics []string

var siteCases = []struct {
	siteName       string
	domain         string
	onHTMLSelector string
	/**
	type HTMLElement struct {
		// Name is the name of the tag
		Name string
		Text string

		// Request is the request object of the element's HTML document
		Request *Request
		// Response is the Response object of the element's HTML document
		Response *Response
		// DOM is the goquery parsed DOM object of the page. DOM is relative
		// to the current HTMLElement
		DOM *goquery.Selection
		// Index stores the position of the current element within all the elements matched by an OnHTML callback
		Index int
		// contains filtered or unexported fields
	}
	**/
	onHTMLFunc func(*colly.HTMLElement)
}{
	{
		siteName:       "U.S. Trade Representative",
		domain:         "https://ustr.gov",
		onHTMLSelector: "ul.listing",
		onHTMLFunc: func(e *colly.HTMLElement) {
			e.DOM.
				ChildrenFiltered("li").
				ChildrenFiltered("a[href]").
				Each(func(_ int, selection *goquery.Selection) {
					if val, exists := selection.Attr("href"); exists {
						articleLinks = append(articleLinks, "https://ustr.gov"+val)
					} else {
						log.Printf("No href found at %s", selection.Text())
					}
				})
		},
	},
	{
		siteName:       "U.S. State Department",
		domain:         "https://www.state.gov/",
		onHTMLSelector: "ul.collection-results",
		onHTMLFunc: func(e *colly.HTMLElement) {
			e.DOM.
				ChildrenFiltered("a.collection-result__link").
				Each(func(_ int, selection *goquery.Selection) {
					if hrefLink, exists := selection.Attr("href"); exists {
						articleLinks = append(articleLinks, hrefLink)
					} else {
						log.Printf("No href found at %s\n", selection.Text())
					}
				})
		},
	},
	{
		siteName:       "U.S. White House",
		domain:         "https://www.whitehouse.gov",
		onHTMLSelector: "div.article-wrapper",
		onHTMLFunc: func(e *colly.HTMLElement) {
			e.DOM.
				ChildrenFiltered("a.news-item__title").
				Each(func(_ int, selection *goquery.Selection) {
					if hrefLink, exists := selection.Attr("href"); exists {
						articleLinks = append(articleLinks, hrefLink)
					} else {
						log.Printf("No href found at %s\n", selection.Text())
					}
				})
		},
	},
	/**
	{
		siteName:            "",
		domain:            "",
		onHTMLSelector: "",
		onHTMLFunc: func(e *colly.HTMLElement) {

		},
	},
	*/
}

func init() {
	links = make([]string, 0)
	if f, err := os.Open("gov_links.txt"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		scanner := bufio.NewScanner(f)
		regx, err := regexp.Compile("") // TODO Match for page counts at the end
		if err != nil {
			log.Fatalf("Error compiling regular expression: %s\n", err.Error())
		}
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if regx.MatchString(line) {
				// TODO fill links
			} else {
				links = append(links, line)
			}
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf(err.Error())
		}
	} else {
		log.Fatalf(err.Error())
	}

	topics = make([]string, 0)
	if f, err := os.Open("gov_links.txt"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			topics = append(topics, strings.TrimSpace(scanner.Text()))
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf(err.Error())
		}
	} else {
		log.Fatalf(err.Error())
	}
	articleLinks = make([]string, 0)
}

func main() {
	c := colly.NewCollector(colly.AllowedDomains(
		"https://www.whitehouse.gov",
		"https://www.state.gov/",
		"https://ustr.gov"))

	for _, siteCase := range siteCases {
		// Find and visit all links
		c.OnHTML(siteCase.onHTMLSelector, siteCase.onHTMLFunc)

		c.OnRequest(func(r *colly.Request) {
			fmt.Printf("Visiting %s at %s\n", siteCase.siteName, r.URL)
		})

		for _, siteUrl := range links {
			if strings.HasPrefix(siteUrl, siteCase.domain) {
				if err := c.Visit(siteUrl); err != nil {
					fmt.Printf(
						"Error while visiting %s at %s: %s\n",
						siteCase.siteName,
						siteUrl,
						err.Error())
				}
			}
		}
	}

	log.Println(articleLinks)
}
