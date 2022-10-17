Dockerfile finder
=======================
Scans input repositories, using github.com's "tree" API endpoint, for files named "Dockerfile" and, when found:
 - greps for "FROM" lines
 - parses out the line's docker [image]:[tag]
 - outputs JSON 

Expects a max of "a few hundred" repositories. Larger datasets may hit github API limits.

The app requires a URL to run. This can be passed as an env variabled named "REPOSITORY_LIST_URL", or as
a command line switch -i [URL]. 

Likewise, it takes an optional github token which can be passed as "GH_TOKEN" or via "-t [TOKEN]".

Build
------------
`go build -o dockerfile-finder main.go`

Run
------------
`REPOSITORY_LIST_URL="http://localhost/urls.txt" GH_TOKEN='githubToken' ./dockerfile-finder`

Usage
------------
`./dockerfile-finder [url] -t [GITHUB_TOKEN] -i [URL]`

`url` expecting a URL holding space-separated data formated as: [github_url] [sha_hash]. See source.txt. 

`GITHUB_TOKEN` (optional) expects a valid github token, used for making api calls

Note
-----------
We do not have reliable methods for dynamically decoding content. Parts of the github API supply encoding information, for example 'base64' as of oct 11, 2022. If they change to base65, our tool would fail. Therefore, we are using decoded data available over the public url https://raw.githubusercontent.com for some "API" calls. 

Reference
-----------
 - https://docs.github.com/en/rest/git/trees#get-a-tree

TODO
-----------
 - increase test coverage
 - clean code
 - incorporate API response headers "x-ratelimit-limit", "x-ratelimit-remaining", and "x-ratelimit-reset"
 - make getData generic by removing json.Unmarshal
