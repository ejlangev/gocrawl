package main

import (
  "fmt"
  "flag"
  "github.com/PuerkitoBio/goquery"
  "net/url"
  // "runtime"
)

var rootUrl *url.URL
var crawledMap map[string]bool
var linkMap map[string]*Page

type Page struct {
  Url string
  HtmlBody *goquery.Document
  Links []string
  Assets []string
}

func main() {
  // Used to control the number of worker
  // goroutines making http requests
  var concurrentRequests int
  // Buffered channel to hold the remaining urls
  // to parse
  urlQueue := make(chan string, 10)
  // Buffered channel to hold the response bodies
  // that still need to be parse
  resultsQueue := make(chan *Page, 10)
  // Channel to deal with completing the crawl
  finishedChannel := make(chan bool)
  // Set up the crawled map
  crawledMap = make(map[string]bool)
  crawledMap["http://joingrouper.com"] = true
  // Set up the linkMap
  linkMap = make(map[string]*Page)
  // Set the root url
  rootUrl, _ = url.Parse("http://joingrouper.com")

  flag.IntVar(
    &concurrentRequests,
    "concurrency",
    4,
    "Number of concurrent requests to perform",
  )
  // Parse command line variables
  flag.Parse()
  // Create the proper number of concurrent workers
  // for consuming off the urlQueue
  fmt.Printf("Concurrency: %d\n", concurrentRequests)
  // Push the initial url into the queue
  urlQueue <- "http://joingrouper.com"

  for i := 0; i < concurrentRequests; i++ {
    go FetchUrls(&urlQueue, &resultsQueue)
  }
  // Start GOMAXPROCS concurrent parsing routines
  for i := 0; i < 1; i++ {
    go ParseResults(&urlQueue, &resultsQueue)
  }

  <- finishedChannel
  fmt.Printf("Hello, world.\n")
}

func FetchUrls(urlQueue *chan string, resultsQueue *chan *Page) {
  // Variable to hold the current url to parse
  var currentUrl string
  for {
    select {
      case currentUrl = <- *urlQueue:
        // Found a url to parse, make a request to find it
        // and push the results into the resultsQueue
        fmt.Printf("Loading HTML for %s\n", currentUrl)
        resp, err := goquery.NewDocument(currentUrl)
        if err != nil {
          HandleFetchError(err)
        } else {
          *resultsQueue <- &Page{
            Url : currentUrl,
            HtmlBody : resp,
          }
        }
    }
  }
}

func ParseResults(urlQueue *chan string, resultsQueue *chan *Page) {
  // Variable to hold the current body off the results queue= <
  var currentPage *Page

  for {
    select {
      case currentPage = <- *resultsQueue:
        // Loop over all the resulting anchor tags
        fmt.Println("Pulled off document to parse")
        currentPage.HtmlBody.Find("a").Each(func(i int, s *goquery.Selection) {
          href, _ := s.Attr("href")

          url, _ := url.Parse(href)

          if needToCrawl(url) {
            // Add this into the list of links to be crawled,
            // filling in any blank data
            if url.Host == "" {
              url.Host = rootUrl.Host
              url.Scheme = rootUrl.Scheme
            }
            // Add this link to the page
            currentPage.Links = append(currentPage.Links, url.String())
            // Add this url to the crawled map
            crawledMap[url.RequestURI()] = true
            fmt.Printf("Adding %s to be crawled\n", url.String())
            *urlQueue <- url.String()
          }
          // Set the attributes of this page and add it to
          // the map

        })
        // Add src tags as assets
        currentPage.HtmlBody.Find("script").Each(func(i int, s *goquery.Selection) {
          src, _ := s.Attr("src")

          if src != "" {
            currentPage.Assets = append(currentPage.Assets, src)
          }
        })
        // Add link tags as static

        printPage(currentPage)
    }
  }
}

func printPage(page *Page) {
  fmt.Printf("URL: %s\n", page.Url)
  fmt.Println("Links:")
  for i := range page.Links {
    fmt.Printf("\t- %s\n", page.Links[i])
  }
  fmt.Println("Assets:")
  for j := range page.Assets {
    fmt.Printf("\t- %s\n", page.Assets[j])
  }
}

func needToCrawl(path *url.URL) bool {
  // Check if this is an absolute path or on the root
  // domain
  if (path.Host == "" || path.Host == rootUrl.Host) &&
     (path.Scheme == "http://" || path.Scheme == "") {
    // Hosts look good, make sure we haven't already
    // crawled this url
    return !crawledMap[path.RequestURI()]
  }

  return false
}

func HandleFetchError(err error) {

}

func HandleParseError(err error) {

}