package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
)

const Version = "1.0.0"

type CliOptions struct {
	ConfigFile       string
	Debug            bool
	Concurrency      int
	Version          bool
	RawInjections    string
	RegexMatch       string
	Replacement      string
	AppendMode       bool
	DecodedParams    bool
	IncludeUnaltered bool
}

type Config struct {
	Rules      map[string]MatchReplaceRule `mapstructure:"rules"`
	Injections []string
	DumbMode   bool
}

func verifyFlags(options *CliOptions) error {
	flag.BoolVar(&options.Debug, "debug", false, "Debug/verbose mode to print more info for failed/malformed URLs or requests")

	flag.IntVar(&options.Concurrency, "w", 15, "Set the concurrency/worker count")
	flag.IntVar(&options.Concurrency, "workers", 15, "Set the concurrency/worker count")

	flag.BoolVar(&options.DecodedParams, "d", false, "Inject with URL decoded params (default is encoded)")
	flag.BoolVar(&options.DecodedParams, "decode", false, "Inject with URL decoded params (default is encoded)")

	flag.BoolVar(&options.Version, "v", false, "Get the current version of qsinject")
	flag.BoolVar(&options.Version, "version", false, "Get the current version of qsinject")

	flag.StringVar(&options.RegexMatch, "m", "", "Regex string to match query string values to be replaced")
	flag.StringVar(&options.RegexMatch, "match", "", "Regex string to match query string values to be replaced")

	flag.StringVar(&options.Replacement, "r", "", "Replacement values (injection) for the matched regex value")
	flag.StringVar(&options.Replacement, "replace", "", "Replacement values (injection) for the matched regex value")

	flag.StringVar(&options.RawInjections, "i", "", "Injections (comma separated) to inject for all query strings")
	flag.StringVar(&options.RawInjections, "injections", "", "Injections (comma separated) to inject for all query strings")

	flag.BoolVar(&options.AppendMode, "a", false, "Append injections to the original query string value (i.e. q=1 > q=1injection")
	flag.BoolVar(&options.AppendMode, "append", false, "Append injections to the original query string value (i.e. q=1 > q=1injection")

	flag.BoolVar(&options.IncludeUnaltered, "iu", false, "Included unaltered URLs in results (for when match and replace doesn't affect that URL)")
	flag.BoolVar(&options.IncludeUnaltered, "include-unaltered", false, "Included unaltered URLs in results (for when match and replace doesn't affect that URL)")

	flag.StringVar(&options.ConfigFile, "c", "", "Pass a regex rules config file instead of flags, which also supports multiple rules per run")
	flag.StringVar(&options.ConfigFile, "config", "", "Pass a regex rules config file instead of flags, which also supports multiple rules per run")

	flag.Parse()

	if options.Version {
		fmt.Println("qsinject version: " + Version)
		os.Exit(0)
	}

	if options.RawInjections != "" {
		config.Injections = strings.Split(options.RawInjections, ",")
	}

	if (options.RegexMatch != "" && options.Replacement == "") || (options.RegexMatch == "" && options.Replacement != "") {
		fmt.Println("If -match or -replace is set, the other must be set as well")
		os.Exit(1)
	}

	if options.RegexMatch != "" {
		config.DumbMode = false
		re, err := regexp.Compile(opts.RegexMatch)
		if err != nil {
			fmt.Println("Regex failed to compile: ", err)
			os.Exit(1)
		}
		rules = append(rules, MatchReplaceRule{Re: re, Replacement: opts.Replacement})
	} else if options.ConfigFile != "" {
		config.DumbMode = false
	} else {
		config.DumbMode = true
	}

	return nil
}

func loadConfig(configFile string) error {
	// In order to ensure dots (.) are not considered as delimiters, set delimiter
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))

	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(&config); err != nil {
		return err
	}

	if err := v.UnmarshalKey("rules", &config); err != nil {
		return err
	}

	if len(config.Rules) > 0 {
		for rule := range config.Rules {
			r := config.Rules[rule]
			re, err := regexp.Compile(r.RawRegex)
			if err != nil {
				fmt.Println("Regex failed to compile: ", err)
				os.Exit(1)
			}
			r.Re = re
			rules = append(rules, r)
		}
	}

	return nil
}
