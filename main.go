package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

var articleLinksOnTopic []string
var articleLinks []string
var links []string
var topics map[string]bool

var wordSplit *regexp.Regexp

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

	articleOnHTMLSelector string
	articleOnHTMLFunc     func(*colly.HTMLElement)
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
		articleOnHTMLSelector: "",
		articleOnHTMLFunc: func(h *colly.HTMLElement) {

		},
	},
	/*{
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
		articleOnHTMLSelector: "",
		articleOnHTMLFunc: func(h *colly.HTMLElement) {

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
		articleOnHTMLSelector: "section.body-content",
		articleOnHTMLFunc: func(e *colly.HTMLElement) {
			e.DOM.
				ChildrenFiltered("p:not(.has-text-align-center)").
				Each(func(_ int, selection *goquery.Selection) {
					parText := selection.Text()
					words := wordSplit.Split(parText, -1)

					for _, word := range words {
						if val, ok := topics[word]; ok && val {
							articleLinksOnTopic = append(articleLinksOnTopic, e.Request.URL.String())
							break
						}
					}
				})
		},
	},*/
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

func compileLinksOnTopic() {
	for _, siteCase := range siteCases {
		c := colly.NewCollector()

		// Find and visit all links
		c.OnHTML(siteCase.articleOnHTMLSelector, siteCase.articleOnHTMLFunc)

		c.OnError(func(r *colly.Response, err error) {
			log.Fatalf("Status code %d Error: %s\n", r.StatusCode, err.Error())
		})

		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("User-Agent", "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
			fmt.Printf("Visiting %s article at %s\n", siteCase.siteName, r.URL)
		})

		for _, siteUrl := range articleLinks {
			if strings.HasPrefix(siteUrl, siteCase.domain) {
				if err := c.Visit(siteUrl); err != nil {
					fmt.Printf(
						"Error while visiting %s at %s: %s\n",
						siteCase.siteName,
						siteUrl,
						err.Error())
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	saveLinksToFile("article_links_on_topic.txt", articleLinksOnTopic)
}

func compileArticleLinks() {
	for _, siteCase := range siteCases {
		c := colly.NewCollector()

		// c.SetRequestTimeout(30 * time.Second)

		// Find and visit all links
		c.OnHTML(siteCase.onHTMLSelector, siteCase.onHTMLFunc)

		c.OnError(func(r *colly.Response, err error) {
			log.Fatalf("Status code %d Error: %s\n", r.StatusCode, err.Error())
		})

		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("User-Agent", "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
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
				time.Sleep(1 * time.Second)
			}
		}
	}

	saveLinksToFile("article_links.txt", articleLinks)
}

func loadArticleLinks() {
	if f, err := os.Open("article_links.txt"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			articleLinks = append(articleLinks, strings.TrimSpace(scanner.Text()))
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf(err.Error())
		}
	} else {
		log.Fatalf(err.Error())
	}
}

func loadListLinks() {
	if f, err := os.Open("gov_links.txt"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		scanner := bufio.NewScanner(f)
		digitRegx, err := regexp.Compile(`[\d]+`) // TODO Match for page counts at the end
		if err != nil {
			log.Fatalf("Error compiling regular expression: %s\n", err.Error())
		}
		regx, err := regexp.Compile(`{page/[\d]+-[\d]+/}$`)
		if err != nil {
			log.Fatalf("Error compiling regular expression: %s\n", err.Error())
		}
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if regx.MatchString(line) {
				location := digitRegx.FindStringIndex(line)
				startIndex := location[0]
				endIndex := location[1]
				pageStart, err := strconv.Atoi(digitRegx.FindString(line[startIndex:]))
				if err != nil {
					log.Fatal("Failed converting index")
				}
				pageEnd, err := strconv.Atoi(digitRegx.FindString(line[endIndex+1:]))
				if err != nil {
					log.Fatal("Failed converting index")
				}

				root := line[:regx.FindStringIndex(line)[0]]

				for i := pageStart; i <= pageEnd; i++ {
					full_link := root + "page/" + strconv.Itoa(i) + "/"
					log.Printf("Created link: %s\n", full_link)
					links = append(links, full_link)
				}
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

	log.Println(len(links))
}

func loadTopics() {
	if f, err := os.Open("topics.txt"); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			topics[strings.TrimSpace(scanner.Text())] = true
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf(err.Error())
		}
	} else {
		log.Fatalf(err.Error())
	}
}

func saveLinksToFile(path string, links []string) {
	if f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err == nil {
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatalf(err.Error())
			}
		}()

		fileWriter := bufio.NewWriter(f)
		for _, link := range links {
			_, err = fileWriter.WriteString(fmt.Sprintf("%s\n", link))
			if err != nil {
				log.Fatalf("Unable to save link: %s, %s\n", link, err.Error())
			}
		}
	} else {
		log.Fatal("Error writing article links to file.")
	}
}

func init() {
	var err error
	wordSplit, err = regexp.Compile(`[ .,!?'[:]-\n\t]+`)
	if err != nil {
		log.Fatalf("")
	}
	links = make([]string, 0)
	loadListLinks()

	topics = make(map[string]bool)
	loadTopics()

	articleLinks = make([]string, 0)
	// loadArticleLinks()

	articleLinksOnTopic = make([]string, 0)
}

func main() {
	compileArticleLinks()
	compileLinksOnTopic()
}
