// main_test.go
package main

import (
	"testing"
	"strings"
	"fmt"
)

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

//func LinesFromReader(r io.Reader) []string {
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
//func TestGetData(t *testing.T) {
//		}

//func getPaths(url string, c chan<- []string) {
//func TestGetPaths(t *testing.T) {
//		}

//func parseLine(line []string) (url, hash, username, repo string) {
//func TestParseLine(t *testing.T) {
//		}

//func findImages(line string, regexp *regexp.Regexp) (imageName string) {
//func TestFindImages(t *testing.T) {
//		}

//func assembleDataStruct(line []string, ch3 chan<- Data) {
//func TestAssembleDataStruct(t *testing.T) {
//		}

// BENCHMARKS
//func BenchmarkTestAssembleDataStruct(b *testing.B) {
//	for i := 0; i<= 1; i++ {
//	}
//}
