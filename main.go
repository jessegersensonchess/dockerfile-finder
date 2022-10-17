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
	data    []Data
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

func removeDockerfileComments(line string) string {
	// removes Dockerfile comments, things like: #FROM image:latest
	substrings := strings.Split(line, "#")
	return substrings[0]
}

func hasContent(str string) bool {
	return len(str) > 0
}

func hasToken(Token string) bool {
	return len(Token) > 0
}

func isWellFormattedData(line []string) bool {
	return len(line) == 2
}

func UrlToLines(url string, c chan<- []string) {
	// gets url and sticks each of it lines into a slice
	resp, err := http.Get(url)
	if err != nil {
		log.Println("ERROR: problem getting url=", url, ", error=", err)
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
	if hasToken(Token) {
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

func findImages(line string) (imageName string) {
	subStrings := strings.Fields(line)
	if len(subStrings) == 0 {
		return
	}
	if strings.ToUpper(subStrings[0]) == "FROM" {
		// handles FROM's optional argument "--platform" (https://docs.docker.com/engine/reference/builder/#from)
		// expected formats:
		// FROM --platform=[platform] ubuntu:latest
		// FROM ubuntu:latest
		if strings.Count(line, "--platform=") > 0 {
			imageName = subStrings[2]
		} else {
			imageName = subStrings[1]
		}
	}
	return imageName
}

func assembleDataStruct(line []string, ch3 chan<- Data) {
	ch, ch2 := make(chan []string), make(chan []string)
	trees := []Tree{}
	url, hash, username, repo := parseLine(line)

	go getPaths(url, ch2)
	paths := <-ch2

	// loop over paths
	for _, path := range paths {
		// Use https://raw.githubusercontent.com to retrive content. The alternate,
		// using the API, requires dynamically decoding data based on the supplied encoding (base64 as of oct 11 2022)
		// We don't have methods for that, so using the decoded content from raw...
		template := "%s/%s/%s/%s/%s"
		dataUrl := fmt.Sprintf(template, githubRawUrl, username, repo, hash, path)

		go UrlToLines(dataUrl, ch)
		linesFromFile := <-ch
		images := extractImageFromDockerfile(linesFromFile)
		tree := Tree{Path: path, URL: dataUrl, Image: images}
		trees = append(trees, tree)
	}
	ch3 <- Data{Sha: line[1], URL: line[0], Tree: trees}
}

func extractImageFromDockerfile(linesFromFile []string) (images []string) {
	// loop over each line in the Dockerfile
	for _, line := range linesFromFile {
		line = removeDockerfileComments(line)
		imageName := findImages(line)
		if hasContent(imageName) {
			images = append(images, imageName)
		}
	}
	return
}

func formatData(data []Data) (result string) {
	byteArray, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	return string(byteArray)
}

func main() {
	ch3 := make(chan Data)
	defaultUrl, _ := os.LookupEnv("REPOSITORY_LIST_URL")

	// command line arguments
	flag.StringVar(&url, "i", defaultUrl, "URL to txt file. Expected format of text file is: [github.com repository] [commit hash]")
	flag.StringVar(&Token, "t", os.Getenv("GH_TOKEN"), "Github Api Token")
	flag.Parse()

	if hasToken(Token) {
		Token = "Bearer " + Token
	}

	ch4 := make(chan []string)
	go UrlToLines(url, ch4)
	lines := <-ch4

	// loop over INPUT file
	for _, v := range lines {
		line := strings.Split(v, " ")
		if isWellFormattedData(line) {
			go assembleDataStruct(line, ch3)
		} else {
			dataRow = Data{}
		}
	}

	// retrieve results from channel
	for _, v := range lines {
		line := strings.Split(v, " ")
		if isWellFormattedData(line) {
			dataRow = <-ch3
			data = append(data, dataRow)
		}
	}
	// print result
	fmt.Printf("%s", formatData(data))
}
