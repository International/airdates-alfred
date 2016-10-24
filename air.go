package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fate-lovely/go-alfred"
)

var listCcommand = "list_shows"
var refreshCommand = "refresh"
var timeFormat = "Mon, 02 Jan 2006"

var currentTime = time.Now()
var currentYear = strconv.Itoa(currentTime.Year())

// A Show represents an entry for a specific show.
type Show struct {
	NameEntry string
}

type ShowSorter []Show

func (a ShowSorter) Len() int           { return len(a) }
func (a ShowSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ShowSorter) Less(i, j int) bool { return a[i].NameEntry < a[j].NameEntry }

func (s *Show) ShowName() string {
	showNameComponents := strings.Split(s.NameEntry, " ")
	return strings.Join(showNameComponents[0:len(showNameComponents)-1], " ")
}

// DayEntries holds shows for a specific day
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

func obtainDayEntries(doc *goquery.Document) ([]DayEntries, error) {
	daysElements := doc.Find(".day")
	days := make([]DayEntries, daysElements.Size())
	var elementError error

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

func buildShowReleaseDates(dayReleases []DayEntries) map[Show]time.Time {
	releaseMap := make(map[Show]time.Time)

	for _, element := range dayReleases {
		date := element.Date
		for _, show := range element.Shows {
			_, ok := releaseMap[show]
			if !ok {
				if date.After(currentTime) {
					releaseMap[show] = date
				}
			}
		}
	}

	return releaseMap
}

//
// func obtainShowNames(doc *goquery.Document) []Show {
// 	showSet := make(map[string]int)
// 	showEntries := doc.Find(".title")
//
// 	showEntries.Each(func(i int, s *goquery.Selection) {
// 		showName := showName(s.Text())
// 		showSet[showName]++
// 	})
//
// 	shows := make([]Show, len(showSet))
// 	showIndex := 0
//
// 	for key := range showSet {
// 		shows[showIndex] = Show{key}
// 		showIndex++
// 	}
//
// 	return shows
// }

func buildAlfredResponseWithShowNames(input map[Show]time.Time) (string, error) {
	var showKeys []Show

	for key := range input {
		showKeys = append(showKeys, key)
	}

	sort.Sort(ShowSorter(showKeys))

	for _, show := range showKeys {
		showDate := input[show]
		alfred.AddItem(alfred.Item{
			Title:    show.NameEntry,
			Subtitle: showDate.Format(timeFormat),
			Arg:      show.NameEntry,
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

	alfredResp, err := buildAlfredResponseWithShowNames(releaseDates)
	if err != nil {
		log.Fatal(err)
	}

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
		log.Fatalf("need path to file, and command ( list_shows | refresh )")
	}

	filePath := args[0]
	command := args[1]

	if command == listCcommand {
		handleListCommand(filePath)
	} else if command == refreshCommand {
		err := handleRefresh(filePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("unknown command", command)
	}

}
