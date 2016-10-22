package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fate-lovely/go-alfred"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var LIST = "list_shows"
var REFRESH = "refresh"
var timeFormat = "Mon, 02 Jan 2006"

var currentTime = time.Now()
var currentYear = strconv.Itoa(currentTime.Year())

type Show struct {
	Name string
}

type DayEntries struct {
	Date  time.Time
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

func obtainDayEntries(doc *goquery.Document) ([]DayEntries, error) {
	daysElements := doc.Find(".day")
	days := make([]DayEntries, daysElements.Size())
	var elementError error = nil

	daysElements.Each(func(i int, s *goquery.Selection) {
		textAttribute := strings.Replace(s.Find(".date").Text(), ".", "", 1) + " " + currentYear
		date, err := time.Parse(timeFormat, textAttribute)

		if err != nil {
			elementError = err
		}

		titles := s.Find(".title")
		shows := make([]Show, titles.Size())

		dayEntries := DayEntries{Date: date, Shows: shows}

		s.Find(".title").Each(func(titleIndex int, s *goquery.Selection) {
			dayEntries.Shows[titleIndex] = Show{s.Text()}
		})

		days = append(days, dayEntries)
	})

	return days, elementError

}

func buildShowReleaseDates(dayReleases []DayEntries) map[string]time.Time {
	releaseMap := make(map[string]time.Time)

	for _, element := range dayReleases {
		date := element.Date
		for _, show := range element.Shows {
			showName := showName(show.Name)
			_, ok := releaseMap[showName]
			if !ok {
				if date.After(currentTime) {
					releaseMap[showName] = date
				}
			}
		}
	}

	return releaseMap
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

func buildAlfredResponseWithShowNames(input map[string]time.Time) (string, error) {
	for showName, showDate := range input {
		alfred.AddItem(alfred.Item{
			Title:    showName,
			Subtitle: showDate.Format(timeFormat),
			Arg:      showName,
			Icon: alfred.Icon{
				Type: "filetype",
				Path: "public.png",
			},
		})
	}

	return alfred.JSON()
}

func handleListCommand(filePath string) {
	pageContent, err := getPageBody(filePath)
	// doc, err := goquery.NewDocument("http://airdates.tv")
	if err != nil {
		log.Fatal(err)
	}

	reader := strings.NewReader(pageContent)
	doc, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		log.Fatal(err)
	}

	de, err := obtainDayEntries(doc)

	if err != nil {
		log.Fatal(err)
	}

	releaseDates := buildShowReleaseDates(de)

	// fmt.Println(releaseDates)
	// showNames := obtainShowNames(doc)
	alfredResp, err := buildAlfredResponseWithShowNames(releaseDates)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Println(alfredResp)
}

func handleRefresh(filePath string) error {
	log.Println("refreshing")
	doc, err := goquery.NewDocument("http://airdates.tv")
	if err != nil {
		return err
	}

	fileHandle, err := os.Create(filePath)
	defer fileHandle.Close()

	if err != nil {
		return err
	}

	htmlString, err := doc.Html()
	if err != nil {
		return err
	}
	content := []byte(htmlString)

	_, err = fileHandle.Write(content)
	return err
}

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		log.Fatalf("need path to file, and command")
	}

	filePath := args[0]
	command := args[1]

	if command == LIST {
		handleListCommand(filePath)
	} else if command == REFRESH {
		handleRefresh(filePath)
	}

}
