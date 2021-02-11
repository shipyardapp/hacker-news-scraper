# hacker-news-scraper

## Description

**hacker-news-scraper** is a simple script for fetching posts from [Hacker News](https://news.ycombinator.com/) matching the provided `score` and `words` parameters. Output can be configured to print to the console or written to a file.

Run `go run scrapehn.go -h` for details on the available flag options.

## Contents

- `scrapehn.go` contains the core scraper logic and uses the [Algolia HN](https://hn.algolia.com/) API to power the querying
- `scrapehn.sh` contains the Bash script logic to run the Go file in the Shipyard platform
