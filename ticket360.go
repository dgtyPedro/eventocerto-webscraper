package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Event struct {
	Title     string
	Date      time.Time
	Thumbnail string
	Location  string
	Genre     string
	Website   string
}

var events []Event

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	var spaceCount int

	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
			spaceCount = 0
		} else {
			if spaceCount == 0 {
				rr = append(rr, r)
			}
			spaceCount++
		}
	}
	return strings.Replace(string(rr), "\n", " ", -1)
}

func visitCategory(h *colly.HTMLElement) {
	parent := h.DOM.Parent()
	items := parent.Find("li")
	items.Each(func(index int, item *goquery.Selection) {
		link := item.Find("a").First()
		href := link.AttrOr("href", "")
		h.Request.Visit(href)
	})
}

func getEvent(h *colly.HTMLElement) Event {
	var event Event

	e := colly.NewCollector(
		colly.AllowedDomains("www.ticket360.com.br"),
	)
	link := h.DOM.Find("a").First()
	href := link.AttrOr("href", "")
	image := h.DOM.Find(".m-portlet img").First()
	src := image.AttrOr("src", "")

	e.OnHTML(".m-section__content", func(h *colly.HTMLElement) {
		title := strings.Replace(h.DOM.Find(".media-heading").First().Text(), "  ", "", -1)
		location := removeSpace(h.DOM.Find(".col-lg-4.notranslate").First().Text())
		date, err := ParseDate(strings.Replace(h.DOM.Find("p").First().Text(), "  ", "", -1))
		if err != nil {
			panic(err)
		}
		fmt.Println("Event finded:", title, date)

		event = Event{
			Title:     title,
			Date:      *date,
			Thumbnail: src,
			Location:  location,
			Genre:     "",
			Website:   "Ticket360",
		}
	})

	e.Visit("https://www.ticket360.com.br/" + href)

	return event
}

func main() {
	fmt.Println("Hello world :)")

	c := colly.NewCollector(
		colly.AllowedDomains("www.ticket360.com.br"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		)
	})

	c.OnHTML("a[title='Próxima Página']", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
	})

	c.OnHTML("a[href='/categoria/1/musica']", func(h *colly.HTMLElement) {
		visitCategory(h)
		fmt.Println("Visiting:", h.Request.URL)
	})

	c.OnHTML(".m-portlet", func(h *colly.HTMLElement) {
		event := getEvent(h)
		events = append(events, event)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error: ", err.Error())
	})

	c.Visit("https://www.ticket360.com.br/")

	file, _ := json.MarshalIndent(events, "", " ")

	_ = ioutil.WriteFile("events.json", file, 0644)

}
