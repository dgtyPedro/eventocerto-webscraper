package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Event struct {
	Title     string
	Link 	  string
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

func removeDuplicates(s []Event) []Event {
	bucket := make(map[Event]bool)
	var result []Event
	for _, event := range s {
		if _, ok := bucket[event]; !ok {
			bucket[event] = true
			result = append(result, event)
		} 
	}
	log.Printf("%s duplicated events", strconv.Itoa(len(s) - len(result)))
	return result
}

func visitCategories(h *colly.HTMLElement, wg *sync.WaitGroup) {
	parent := h.DOM.Parent()
	items := parent.Find("li")
	items.Each(func(index int, item *goquery.Selection) {
		link := item.Find("a").First()
		href := link.AttrOr("href", "")
		h.Request.Visit(href)
	})
}

func getEvent(h *colly.HTMLElement, wg *sync.WaitGroup, currentCategory string) {
	wg.Add(1)
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

		event = Event{
			Title:     title,
			Link:      "https://www.ticket360.com.br/" + href,
			Date:      *date,
			Thumbnail: src,
			Location:  location,
			Genre:     currentCategory,
			Website:   "Ticket360",
		}
		events = append(events, event)
		wg.Done()
	})

	go e.Visit("https://www.ticket360.com.br/" + href)

}

func main() {
	start := time.Now()
	log.Printf("Hello world :)")
	wg := &sync.WaitGroup{}

	db, err := dbConnect()
	if err != nil {
		log.Printf("Error %s when getting db connection", err)
		return
	}
	defer db.Close()
	log.Printf("Successfully connected to database")

	c := colly.NewCollector(
		colly.AllowedDomains("www.ticket360.com.br"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		)
	})

	c.OnHTML("a[href='/categoria/1/musica']", func(h *colly.HTMLElement) {
		visitCategories(h, wg)
	})

	c.OnHTML("a[title='Próxima Página']", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
	})

	c.OnHTML(".m-portlet", func(h *colly.HTMLElement) {
		body := h.Response.Body
		reader := bytes.NewReader(body)
		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			log.Fatal(err)
		}
		currentCategory := doc.Find(".m-subheader__title").Text()
		if currentCategory != "" {
			getEvent(h, wg, currentCategory)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %s", err.Error())
	})

	c.Visit("https://www.ticket360.com.br/")

	wg.Wait()

	events := removeDuplicates(events)

	file, _ := json.MarshalIndent(events, "", " ")

	_ = ioutil.WriteFile("events.json", file, 0644)

	err = resetWebsiteEvents(db, "Ticket360")
	if err != nil {
		log.Printf("Reset Ticket360 events failed with error %s", err)
	}

	log.Printf("%s events founded", strconv.Itoa(len(events)))

	for _, event := range events {
		err = insertEvent(db, event)
		if err != nil {
			log.Printf("Insert event failed with error %s", err)
		}
	}

	log.Printf("Completed in: %s", time.Since(start))
}
