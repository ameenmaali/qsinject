package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
)

type MatchReplaceRule struct {
	RawRegex    string `mapstructure:"regex"`
	Replacement string `mapstructure:"replacement"`
	AppendMode  bool   `mapstructure:"append"`
	Re          *regexp.Regexp
}

func getRegexReplacedUrls(u *url.URL, results chan string) error {
	queryStrings, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		for qs, values := range queryStrings {
			for index, val := range values {
				replacedValue := rule.MatchAndReplace(queryStrings[qs][index])
				// If there is no match, continue unless include unaltered flag is enabled
				if replacedValue == val && !opts.IncludeUnaltered {
					continue
				}

				queryStrings[qs][index] = replacedValue
				// Append values only if the replaced value isn't the original (i.e. a match was made)
				if (opts.AppendMode || rule.AppendMode) && val != replacedValue {
					queryStrings[qs][index] = val + replacedValue
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

func (r *MatchReplaceRule) MatchAndReplace(value string) string {
	return r.Re.ReplaceAllString(value, r.Replacement)
}
