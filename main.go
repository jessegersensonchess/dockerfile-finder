// outputs JSON describing Dockerfile(s) found in the input github repositories
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	githubRawUrl = "https://raw.githubusercontent.com"
	githubApi    = "https://api.github.com"
	Filename     = "Dockerfile"
)

var (
	// github token
	Token string
	// input URL
	url     string
	dataRow Data
)

type apiResponse struct {
	Sha  string `json:"sha"`
	URL  string `json:"url"`
	Tree []struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		Sha  string `json:"sha"`
		Size int    `json:"size,omitempty"`
		URL  string `json:"url"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}

type Data struct {
	Sha  string `json:"sha"`
	URL  string `json:"url"`
	Tree []Tree `json:"tree"`
}

type Tree struct {
	Path  string   `json:"path"`
	Size  int      `json:"size,omitempty"`
	URL   string   `json:"url"`
	Image []string `json:"image"`
}

func UrlToLines(url string, c chan<- []string) {
	// gets url and sticks each of it lines into a slice
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("FATAL: problem getting url=", url, ", error=", err)
	}
	defer resp.Body.Close()
	c <- LinesFromReader(resp.Body)
}

func LinesFromReader(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func getData(url string) (apiResponse, error) {
	// gets url and unmarshals it in an apiResponse
	c := apiResponse{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	if len(Token) > 0 {
		//log.Println("setting Authorization header")
		req.Header.Set("Authorization", Token)
	}
	if err != nil {
		return c, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return c, err
	}
	if res.StatusCode != 200 {
		log.Fatal("FATAL: HTTP response was not 200. StatusCode=", res.StatusCode, ", url=", url, ", message=Confirm token is valid")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func getPaths(url string, c chan<- []string) {
	result := []string{}
	apiResponse, err := getData(url)
	if err != nil {
		log.Fatal("ERROR:", err)
	}
	if apiResponse.Truncated == true {
		log.Fatal("FATAL: action=aborting, reason=api response was truncated, likely because there are over 100000 files in the repository. We can not confirm we've found all the files we are seeking")
	}
	for _, i := range apiResponse.Tree {
		if strings.HasSuffix(i.Path, Filename) {
			result = append(result, i.Path)
		}
	}
	c <- result
}

func parseLine(line []string) (url, hash, username, repo string) {
	hash = line[1]
	path := strings.Split(line[0], "/")
	username = path[3]
	repo = strings.Replace(path[4], ".git", "", 1)
	// "trees" endpoint is similar to find ./ -- see https://docs.github.com/en/rest/git/trees#get-a-tree
	template := "%s/repos/%s/%s/git/trees/%s?recursive=1"
	url = fmt.Sprintf(template, githubApi, username, repo, hash)
	return url, hash, username, repo
}

func findImages(line string, regexp *regexp.Regexp) (imageName string) {
	lineCleaned := regexp.ReplaceAllString(line, "")
	isFrom := false
	if len(strings.Fields(lineCleaned)) > 0 {
		if strings.ToUpper(strings.Fields(lineCleaned)[0]) == "FROM" {
			isFrom = true
		}
	}
	if isFrom == true {
		subString := strings.Fields(lineCleaned)
		// handles FROM's optional argument "--platform" (https://docs.docker.com/engine/reference/builder/#from)
		// expected formats:
		// FROM --platform=[platform] ubuntu:latest
		// FROM ubuntu:latest
		if strings.Count(lineCleaned, "--platform=") > 0 {
			imageName = subString[2]
		} else {
			imageName = subString[1]
		}
	}
	return imageName
}

func assembleDataStruct(line []string, ch3 chan<- Data) {
	ch, ch2 := make(chan []string), make(chan []string)
	trees := []Tree{}
	url, hash, username, repo := parseLine(line)

	// regex removes Dockerfile comments to simplify string matching
	// i.e. removes things like #FROM image:latest
	regexp, _ := regexp.Compile(`#.*`)

	go getPaths(url, ch2)
	paths := <-ch2

	// loop over paths
	for _, i := range paths {
		// Use https://raw.githubusercontent.com to retrive content. The alternate,
		// using the API, requires dynamically decoding data based on the supplied encoding (base64 as of oct 11 2022)
		// We don't have methods for that, so using the decoded content from raw...
		template := "%s/%s/%s/%s/%s"
		dataUrl := fmt.Sprintf(template, githubRawUrl, username, repo, hash, i)

		go UrlToLines(dataUrl, ch)
		fileLine := <-ch

		images := []string{}
		// loop over each line in Dockerfile
		for _, line := range fileLine {
			imageName := findImages(line, regexp)
			if len(imageName) > 0 {
				images = append(images, imageName)
			}
		}
		tree := Tree{Path: i, URL: dataUrl, Image: images}
		trees = append(trees, tree)
	}
	ch3 <- Data{Sha: line[1], URL: line[0], Tree: trees}
}

func main() {
	data := []Data{}
	ch3 := make(chan Data)

	// command line arguments
	defaultUrl, _ := os.LookupEnv("REPOSITORY_LIST_URL")
	flag.StringVar(&url, "i", defaultUrl, "URL to txt file. Expected format of text file is: [github.com repository] [commit hash]")
	flag.StringVar(&Token, "t", os.Getenv("GH_TOKEN"), "Github Api Token")
	flag.Parse()
	if len(Token) > 0 {
		Token = "Bearer " + Token
	}

	ch4 := make(chan []string)
	go UrlToLines(url, ch4)
	lines := <-ch4

	// loop over INPUT file
	for _, v := range lines {
		line := strings.Split(v, " ")
		if len(line) == 2 {
			go assembleDataStruct(line, ch3)
		} else {
			dataRow = Data{}
		}
	}

	// retrieve results from channel
	for _, v := range lines {
		line := strings.Split(v, " ")
		if len(line) == 2 {
			dataRow = <-ch3
			data = append(data, dataRow)
		}
	}

	byteArray, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", string(byteArray))
}
