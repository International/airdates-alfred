package main

import (
	// "fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Show struct {
	Name string
}

func getPageBody(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func obtainShowNames(doc *goquery.Document) []Show {
	showSet := make(map[string]int)
	showEntries := doc.Find(".title")

	showEntries.Each(func(i int, s *goquery.Selection) {
		showNameComponents := strings.Split(s.Text(), " ")
		showName := strings.Join(showNameComponents[0:len(showNameComponents)-1], " ")
		showSet[showName] += 1
	})

	shows := make([]Show, len(showSet))
	showIndex := 0

	for key, _ := range showSet {
		shows[showIndex] = Show{key}
		showIndex += 1
	}

	return shows
}

func main() {
	pageContent, err := getPageBody("airdates.html")
	// doc, err := goquery.NewDocument("http://airdates.tv")
	if err != nil {
		log.Fatal(err)
	}

	reader := strings.NewReader(pageContent)
	doc, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		log.Fatal(err)
	}

	for _, show := range obtainShowNames(doc) {
		log.Println(show.Name)
	}

}
