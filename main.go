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
	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)

	log.Println("Serving at http://localhost:80")
	err := http.ListenAndServe(":80", mux)
	log.Fatal(err)
}

func Index(w http.ResponseWriter, r *http.Request) {
	c := colly.NewCollector(
		colly.AllowedDomains("www.wired.com", "wired.com"),
	)

	c.OnHTML("a[href] h2", func(e *colly.HTMLElement) {
		parentA := e.DOM.Parent()

		if parentA.Is("a") {
			article := Article{Title: e.Text, Link: ""}
			href, exists := parentA.Attr("href")
			var sb strings.Builder
			if exists {
				// Ensure the link is absolute
				if len(href) > 0 && href[0] == '/' {
					sb.WriteString("https://www.wired.com")
				}
			} else {
				article.Title = ""
				log.Printf("Warning: Article with title %s does not have a link.", article.Title)
			}
			sb.WriteString(href)
			article.Link = sb.String()

			articles = append(articles, article)
		}

	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with error:", err)
	})

	c.Visit("https://www.wired.com")

	// Sort in reverse chronological order.
	sort.Slice(articles, func(i, j int) bool {
		return i > j
	})
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
}
