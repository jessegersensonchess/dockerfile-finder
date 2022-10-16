// main_test.go
package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestHasTokenFalse(t *testing.T) {
	Token := ""
	expecting := false
	if hasToken(Token) {
		t.Errorf("expecting: %v but got %v", expecting, hasToken(Token))
	}
}
func TestHasTokenTrue(t *testing.T) {
	Token := "myToken"
	expecting := true
	if hasToken(Token) {
		// this test passed!
	} else {
		t.Errorf("expecting: %v but got %v", expecting, hasToken(Token))
	}
}
func TestisWellFormattedData(t *testing.T) {
	arr := []string{"a"}
	expecting := false
	if isWellFormattedData(arr) {
		t.Errorf("expecting: %v but got %v", expecting, isWellFormattedData(arr))
	}
}

func TestUrlToLines(t *testing.T) {
	ch := make(chan []string)
	dataUrl := "https://gist.githubusercontent.com/jmelis/c60e61a893248244dc4fa12b946585c4/raw/25d39f67f2405330a6314cad64fac423a171162c/sources.txt"
	go UrlToLines(dataUrl, ch)
	fileLine := <-ch
	if len(fileLine) != 2 {
		t.Errorf("expecting: 2 but got %d", len(fileLine))
	}
	close(ch)
}

func TestLinesFromReader(t *testing.T) {
	r := strings.NewReader("hello world")
	a := fmt.Sprintf("%s", LinesFromReader(r))
	if len(a) != 13 {
		t.Errorf("expecting: 13 but got %d", len(a))
	}
	if a != "[hello world]" {
		t.Errorf("expecting: string with value 'hello world', but got %v, type=%T", a, a)
	}

}

//func getData(url string) (apiResponse, error) {
func TestGetData(t *testing.T) {
	url := "https://api.github.com/repos/app-sre/container-images/git/trees/c260deaf135fc0efaab365ea234a5b86b3ead404?recursive=1"
	resp, _ := getData(url)
	if fmt.Sprintf("%T", resp) != "main.apiResponse" {
		t.Errorf("%T not expecting getData(url)..", resp)
	}

}

//func getPaths(url string, c chan<- []string) {
func TestGetPaths(t *testing.T) {
	url := "https://api.github.com/repos/app-sre/qontract-reconcile/git/trees/30af65af14a2dce962df923446afff24dd8f123e?recursive=1"
	ch := make(chan []string)
	go getPaths(url, ch)
	paths := <-ch
	expecting := "dockerfiles/Dockerfile"
	if paths[0] != expecting {
		t.Errorf("expecting: %s but got %s", expecting, paths[0])
	}
}

func TestParseLine(t *testing.T) {
	line := []string{"https://github.com/app-sre/qontract-reconcile.git", "30af65af14a2dce962df923446afff24dd8f123e"}
	url, hash, username, repo := parseLine(line)
	if url != "https://api.github.com/repos/app-sre/qontract-reconcile/git/trees/30af65af14a2dce962df923446afff24dd8f123e?recursive=1" {
		t.Errorf("expecting url=https://api.github.com/repos/app-sre/qontract-reconcile/git/trees/30af65af14a2dce962df923446afff24dd8f123e?recursive=1 , got %s", url)
	}
	if hash != "30af65af14a2dce962df923446afff24dd8f123e" {
		t.Errorf("expecting hash=30af65af14a2dce962df923446afff24dd8f123e , got %s", hash)
	}
	if username != "app-sre" {
		t.Errorf("expecting username=app-sre, got %s", username)
	}
	if repo != "qontract-reconcile" {
		t.Errorf("expecting: repo= qontract-reconcile, got %s", repo)
	}
}

//func findImages(line string, regexp *regexp.Regexp) (imageName string) {
func TestFindImages(t *testing.T) {
	line := "FROM registry.access.redhat.com/ubi8/go-toolset:latest as builder"
	expected := "registry.access.redhat.com/ubi8/go-toolset:latest"

	if findImages(line) != expected {
		t.Errorf("expecting: expected got %v", findImages(line))
	}
}
func TestFindImagesPlatform(t *testing.T) {
	line := "FROM --platform=x86-64 registry.access.redhat.com/ubi8/go-toolset:latest as builder"
	expected := "registry.access.redhat.com/ubi8/go-toolset:latest"

	if findImages(line) != expected {
		t.Errorf("expecting: expected got %v", findImages(line))
	}
}

//func assembleDataStruct(line []string, ch3 chan<- Data) {
//func TestAssembleDataStruct(t *testing.T) {
//		}

// BENCHMARKS
//func BenchmarkTestAssembleDataStruct(b *testing.B) {
//	for i := 0; i<= 1; i++ {
//	}
//}
