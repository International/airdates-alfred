package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fate-lovely/go-alfred"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var LIST = "list_shows"

type Show struct {
	Name string
}

type DayEntries struct {
	Date  string
	Shows []Show
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

func showName(input string) string {
	showNameComponents := strings.Split(input, " ")
	return strings.Join(showNameComponents[0:len(showNameComponents)-1], " ")

}

func obtainDayEntries(doc *goquery.Document) []DayEntries {
	daysElements := doc.Find(".day")
	days := make([]DayEntries, daysElements.Size())

	daysElements.Each(func(i int, s *goquery.Selection) {
		date := s.Find(".date").Text()

		titles := s.Find(".title")
		shows := make([]Show, titles.Size())

		dayEntries := DayEntries{Date: date, Shows: shows}

		s.Find(".title").Each(func(titleIndex int, s *goquery.Selection) {
			dayEntries.Shows[titleIndex] = Show{s.Text()}
		})

		days = append(days, dayEntries)
	})

	return days

}

func obtainShowNames(doc *goquery.Document) []Show {
	showSet := make(map[string]int)
	showEntries := doc.Find(".title")

	showEntries.Each(func(i int, s *goquery.Selection) {
		showName := showName(s.Text())
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

func buildAlfredResponseWithShowNames(input []Show) (string, error) {
	for _, item := range input {
		alfred.AddItem(alfred.Item{
			Title:    item.Name,
			Subtitle: item.Name,
			Arg:      item.Name,
			Icon: alfred.Icon{
				Type: "filetype",
				Path: "public.png",
			},
		})
	}

	return alfred.JSON()
}

func main() {
	args := os.Args[1:]
	command := ""

	if len(args) != 1 {
		command = LIST
	}

	if command == LIST {
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

		showNames := obtainShowNames(doc)
		alfredResp, err := buildAlfredResponseWithShowNames(showNames)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(alfredResp)

	}

}
