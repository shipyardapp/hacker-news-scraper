package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	scheme       = "http"
	host         = "hn.algolia.com"
	path         = "api/v1/search_by_date"
	defaultQuery = "query=%s&tags=(story,show_hn,ask_hn)&numericFilters=created_at_i>%d"
	filterQuery  = "query=%s&tags=(story,show_hn,ask_hn)&numericFilters=created_at_i>%d,points<=%d"
)

type hit struct {
	ObjectID    string `json:"objectID"`
	Title       string `json:"title"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
}

type hits struct {
	Hits []hit `json:"hits"`
}

func main() {
	wordsString := flag.String("words", "", "(conditional) comma-separate list of words to scrape for")
	wordsFilePath := flag.String("file", "", "(conditional) path to json file with array of words to scrape for")
	score := flag.Int("score", 0, "(optional) retrieve stories with scores equal to or less than the provided value")
	since := flag.Int("since", 24, "(required) retrive stories within the past number of hours")
	output := flag.String("output", "", "(optional) target CSV file to write results to")

	flag.Parse()

	if *wordsString != "" && *wordsFilePath != "" {
		log.Fatal("error: only provide one scrape word source (string or file)")
	} else if *wordsString == "" && *wordsFilePath == "" {
		log.Fatal("error: one scrape word source must be provided (string or file)")
	}

	scrapeWords := []string{}
	if *wordsString != "" {
		scrapeWords = strings.Split(*wordsString, ",")
	} else if *wordsFilePath != "" {
		jsonBytes, err := ioutil.ReadFile(*wordsFilePath)
		if err != nil {
			log.Fatalf("error: reading json file: %s", err.Error())
		}

		var jsonWords []string
		if err := json.Unmarshal(jsonBytes, &jsonWords); err != nil {
			log.Fatalf("error: unmarshalling json file: %s", err.Error())
		}
		scrapeWords = jsonWords
	}

	if len(scrapeWords) == 0 {
		log.Fatal("error: at least one scrape word must be provided")
	}

	apiURL := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	fullOutput := [][]string{
		[]string{"URL", "ID", "TITLE", "POINTS", "COMMENTS"},
	}
	checkIDs := make(map[string]struct{})
	start := time.Now().Add(time.Hour * time.Duration(*since) * -1).Unix()
	for _, scrapeWord := range scrapeWords {
		if *score <= 0 {
			apiURL.RawQuery = fmt.Sprintf(defaultQuery, url.QueryEscape(scrapeWord), start)
		} else {
			apiURL.RawQuery = fmt.Sprintf(filterQuery, url.QueryEscape(scrapeWord), start, *score)
		}

		resp, err := http.Get(apiURL.String())
		if err != nil {
			log.Fatalf("error: calling hn algolia api: %s\n", err.Error())
		}

		hitsJSON := hits{}
		if err := json.NewDecoder(resp.Body).Decode(&hitsJSON); err != nil {
			log.Fatalf("error: decoding hn algolia response body: %s\n", err.Error())
		}

		hitsOutput := [][]string{}
		for _, result := range hitsJSON.Hits {
			url := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", result.ObjectID)

			if _, ok := checkIDs[result.ObjectID]; !ok {
				hitsOutput = append(hitsOutput, []string{
					url,
					result.ObjectID,
					result.Title,
					strconv.Itoa(result.Points),
					strconv.Itoa(result.NumComments),
				})
				checkIDs[result.ObjectID] = struct{}{}
			}
		}

		fullOutput = append(fullOutput, hitsOutput...)
	}

	if *output == "" {
		for _, hitOutput := range fullOutput {
			log.Println(hitOutput)
		}
	} else {
		file, err := os.Create(*output)
		if err != nil {
			log.Fatalf("error: creating csv output: %s", err.Error())
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		for _, hitOutput := range fullOutput {
			if err := writer.Write(hitOutput); err != nil {
				log.Fatalf("error: writing csv output: %s", err.Error())
			}
		}
	}
}
