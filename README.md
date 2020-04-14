# qsfuzz

qsinject (Query String Inject) is a tool that allows you to quickly substitute query string values with regex matches, one-at-a-time. 

Injections are done one-at-a-time for URLs with multiple query strings to ensure requests aren't broken if certain
parameters are relied on. URLs that don't have query strings will be ignored. Much of this logic is ported over from [qsfuzz](https://github.com/ameenmaali/qsfuzz)
to `qsinject` as a standalone tool.

`qsinject` has 2 modes, "dumb" mode, and "regex" mode.
* Dumb mode will allow you to pass in a simple comma separated list of injections that will inject each query string value, one-at-a-time
* Regex mode will allow you to define rules and only replace a query string value if it matches the defined regex

As a side-benefit, `qsinject` does deduplication to remove duplicates of the same URL and query string keys (with only differing values)

## Installation
```
go get github.com/ameenmaali/qsinject
```

## Usage
inject takes URLs (with query strings) from stdin, of which you will most likely want in a file such as:
```
$ cat urls.txt
https://www.google.com/reflectedXss?param=123
https://test.com
http://hello.com/sqli?qs=444&dd=22
https://redir.auth.com/redir?url=https://test.com&redirect_to=http://domain.gov
https://xD.com/redirect?url=https://evil.com
```

Optionally, you can supply `qsinject` with a rule/config file (see `rule-example.yaml` for an example) to perform multiple
regex match/replacements in a given run. If you do not use a config file, you'll be limited to replacing against a single rule.

```yaml
$ cat rules.yaml
rules:
  UrlUpdate:
    regex: '^http(s)?:\/\/.+'
    replacement: 'https://example.net/home'
    append: false
  XSS:
    regex: '^[a-zA-Z0-9]+$'
    replacement: '"><h2>asd</h2>'
    append: true
  SqlInjection:
    regex: '^[a-zA-Z0-9]+$'
    replacement: "'"
    append: true
```

#### Important Notes for Config files

You can have as many rules as you'd like. These are the currently supported fields, annotated with comments above the field:

```yaml
# This should never change, and indicates the start of the rules list
rules:
  # This should be set to the rule's name you are defining
  ruleName:
    # This is the regex (string) value you'd like to match query string values against. Be careful with escaping, recommended to insert in single quotes
    regex: 
    # This is a (string) value for the injection you'd like to insert for the matched regex
    replacement:
    # This is a (boolean) value for whether you want to append the injection after the original value (true), or replace all together (false)
    append:
```

## Help
```
$ qsinject -h
Usage of qsinject:
  -a	
        Append injections to the original query string value (i.e. q=1 > q=1injection
  -append
    	Append injections to the original query string value (i.e. q=1 > q=1injection
  -c string
    	Pass a regex rules config file instead of flags, which also supports multiple rules per run
  -config string
    	Pass a regex rules config file instead of flags, which also supports multiple rules per run
  -d	
        Inject with URL decoded params (default is encoded)
  -decode
    	Inject with URL decoded params (default is encoded)
  -debug
    	Debug/verbose mode to print more info for failed/malformed URLs
  -i string
    	Injections (comma separated) to inject for all query strings
  -injections string
    	Injections (comma separated) to inject for all query strings
  -iu
    	Included unaltered URLs in results (for when match and replace doesn't affect that URL)
  -include-unaltered
    	Included unaltered URLs in results (for when match and replace doesn't affect that URL)
  -m string
    	Regex string to match query string values to be replaced
  -match string
    	Regex string to match query string values to be replaced
  -r string
    	Replacement values (injection) for the matched regex value
  -replace string
    	Replacement values (injection) for the matched regex value
  -v	
        Get the current version of qsinject
  -version
    	Get the current version of qsinject
  -w int
    	Set the concurrency/worker count (default 15)
  -workers int
    	Set the concurrency/worker count (default 15)
```

## Examples

Replace URLs in query string values with your Burp Collaborator instance

`cat urls.txt | qsinject -m "^http(s)?:\/\/.+" -r "https://myinstance.burpcollaborator.net"`

Using a rule config file, match and replace against multiple rules:

`cat urls.txt | qsinject -c rules.yaml`

Replace URLs in "Dumb" mode, injecting a list of query strings one-at-a-time in each query string value:

`cat urls.txt | qsinject -i "val1,val2,val3,val4"`

Replace URLs with a rule config file, include all results (even if unaltered), and include values decoded:

`cat urls.txt | qsinject -c rules.yaml -iu -decode`
