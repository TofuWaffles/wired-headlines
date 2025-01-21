package main

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Article represents a news headline with its link
type Article struct {
	Title string
	Link  string
}

// Slice to store the scraped articles
var articles []Article

func main() {
	// Initialize the Colly collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.wired.com", "wired.com"),
	)

	// Find and scrape headlines and links
	c.OnHTML("a[href] h2", func(e *colly.HTMLElement) {
		// Extract the parent <a> tag
		parentA := e.DOM.Parent()

		// Check if the parent is an <a> tag
		if parentA.Is("a") {
			// Extract the href attribute from the parent <a> tag
			article := Article{Title: e.Text, Link: ""}
			href, exists := parentA.Attr("href")
			var sb strings.Builder
			if exists {
				log.Printf("Link exists for article: %s", e.Text)
				// Ensure the link is absolute
				if len(href) > 0 && href[0] == '/' {
					log.Printf("Appending to link for %s", e.Text)
					sb.WriteString("https://www.wired.com")
				}
			} else {
				// Print the text of the <h2> tag and the href
				article.Title = ""
				log.Printf("Warning: Article with title %s does not have a link.", article.Title)
			}
			sb.WriteString(href)
			article.Link = sb.String()

			articles = append(articles, article)
		}

	})
	// Handle request errors
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with error:", err)
	})

	// Start scraping
	log.Println("Scraping headlines...")
	c.Visit("https://www.wired.com")

	// Sort articles in reverse chronological order (newest first)
	log.Println("Here are the articles:")
	log.Println(articles)
	sort.Slice(articles, func(i, j int) bool {
		return i > j // Reverse order
	})

	// Serve the scraped data on a web page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Wired Headlines</title>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; }
				.container { max-width: 800px; margin: 0 auto; padding: 20px; }
				.article { margin-bottom: 10px; }
				a { text-decoration: none; color: #0073e6; }
				a:hover { text-decoration: underline; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Latest Wired Headlines</h1>
				{{range .}}
					<div class="article">
						<a href="{{.Link}}" target="_blank">{{.Title}}</a>
					</div>
				{{end}}
			</div>
		</body>
		</html>
		`

		t := template.Must(template.New("webpage").Parse(tmpl))
		if err := t.Execute(w, articles); err != nil {
			log.Println("Template execution error:", err)
		}
	})

	log.Println("Serving at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
