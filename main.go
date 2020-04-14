package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
)

type Task struct {
	Url *url.URL
}

var config Config
var opts CliOptions
var rules []MatchReplaceRule

func getUrlsFromFile() ([]*url.URL, error) {
	deduplicatedUrls := make(map[string]bool)
	var urls []*url.URL

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		providedUrl := scanner.Text()
		// Only include properly formatted URLs
		u, err := url.Parse(providedUrl)
		if err != nil {
			continue
		}

		queryStrings := u.Query()

		// Only include URLs that have query strings unless extra params are provided
		if len(queryStrings) == 0 {
			continue
		}

		// Use query string keys when sorting in order to get unique URL & Query String combinations
		params := make([]string, 0)
		for param, _ := range queryStrings {
			params = append(params, param)
		}
		sort.Strings(params)

		key := fmt.Sprintf("%s%s?%s", u.Hostname(), u.EscapedPath(), strings.Join(params, "&"))

		// Only output each host + path + params combination once, regardless if different param values
		if _, exists := deduplicatedUrls[key]; exists {
			continue
		}
		deduplicatedUrls[key] = true

		urls = append(urls, u)
	}
	return urls, scanner.Err()
}

func getInjectedUrls(u *url.URL, results chan string) error {
	// If query strings can't be parsed, set query strings as empty
	queryStrings, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	for _, injection := range config.Injections {
		injection = strings.TrimSpace(injection)
		for qs, values := range queryStrings {
			for index, val := range values {
				queryStrings[qs][index] = injection
				if opts.AppendMode {
					queryStrings[qs][index] = val + injection
				}
				query, err := getInjectedQueryString(queryStrings)
				if err != nil {
					if opts.Debug {
						fmt.Fprintf(os.Stderr, "Error decoding parameters: ", err)
					}
				}
				u.RawQuery = query
				results <- u.String()

				// Set back to original qs val to ensure we only update one parameter at a time
				queryStrings[qs][index] = val
			}
		}
	}
	return nil
}

func getInjectedQueryString(injectedQs url.Values) (string, error) {
	var qs string
	// TODO: Find a better solution to turn the qs map into a decoded string
	decodedQs, err := url.QueryUnescape(injectedQs.Encode())
	if err != nil {
		return "", err
	}

	if opts.DecodedParams {
		qs = decodedQs
	} else {
		qs = injectedQs.Encode()
	}

	return qs, nil
}

func main() {
	err := verifyFlags(&opts)
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	if opts.ConfigFile != "" {
		if err := loadConfig(opts.ConfigFile); err != nil {
			fmt.Println("Failed loading config:", err)
			os.Exit(1)
		}
	}

	urls, err := getUrlsFromFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tasks := make(chan Task)
	results := make(chan string)

	var wg sync.WaitGroup
	var resultWg sync.WaitGroup

	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			for task := range tasks {
				task.execute(results)
			}
			wg.Done()
		}()
	}

	resultWg.Add(1)
	go func() {
		for u := range results {
			fmt.Println(u)
		}
		resultWg.Done()
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for _, u := range urls {
		tasks <- Task{Url: u}
	}

	close(tasks)
	resultWg.Wait()
}

func (t Task) execute(resultsChan chan string) {
	var err error
	if config.DumbMode {
		err = getInjectedUrls(t.Url, resultsChan)
	} else {
		err = getRegexReplacedUrls(t.Url, resultsChan)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}
