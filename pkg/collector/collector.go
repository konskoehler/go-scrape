package collector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
	"github.com/konskoehler/go-scrape/pkg/sale"
)

func SetupQueue(threads int, queueStorage int) queue.Queue {
	consumerThreads := threads
	Q, _ := queue.New(
		consumerThreads,
		&queue.InMemoryQueueStorage{MaxSize: queueStorage},
	)
	return *Q
}

func Run(threads int, queueStorage int, baseUrl string, sales *[]sale.Sale) {

	q := SetupQueue(threads, queueStorage)
	q.AddURL(baseUrl)
	detailQ := SetupQueue(threads, queueStorage)

	c, _ := NewCollector(&detailQ)
	detailC, _ := NewDetailCollector(sales)

	q.Run(c)
	detailQ.Run(detailC)
}

func NewBaseCollector() (*colly.Collector, error) {
	c := colly.NewCollector()

	extensions.RandomUserAgent(c)

	// set Proxy
	//c.SetProxy(proxyAddress)
	return c, nil
}

func NewCollector(Q *queue.Queue) (*colly.Collector, error) {

	c, _ := NewBaseCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		saleURL := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Index(saleURL, "/itm/") != -1 {
			Q.AddURL(saleURL)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Got this error:", e)
	})

	return c, nil
}

func NewDetailCollector(sales *[]sale.Sale) (*colly.Collector, error) {

	c, _ := NewBaseCollector()

	c.OnHTML("div[id=Body]", func(e *colly.HTMLElement) {
		title := e.DOM.Clone().Children().Find("#itemTitle").Children().Remove().End().Text()
		dateRaw := e.DOM.Clone().Find("#bb_tlft").Children().Remove().End().Text()
		dateDays := strings.ReplaceAll(strings.ReplaceAll(dateRaw, "\n", ""), "\t", "")
		dateMinutes := e.ChildText(".endedDate")
		date_ := strings.Replace(dateDays+" "+dateMinutes, "MEZ", "+01", 1)
		layout := "02. Jan. 2006 15:04:05 -07"
		dateFinal, _ := time.Parse(layout, date_)

		cost := strings.Replace(e.ChildText(".notranslate.vi-VR-cvipPrice"), "EUR ", "", 1)
		if cost == "" { // Seller accepted offer
			cost = strings.Replace(e.ChildText("#prcIsum"), "EUR ", "", 1)
		}
		if cost == "" {
			cost = strings.Replace(e.ChildText("#mm-saleDscPrc"), "EUR ", "", 1)
		}

		var proposalAccepted bool
		if e.ChildText(".vi-boLabel") != "" {
			proposalAccepted = true
		} else {
			proposalAccepted = false
		}

		shipping := strings.Replace(e.ChildText("#fshippingCost"), "EUR ", "", 1)
		if shipping == "" {
			shipping = "0"
		}
		url := e.Request.URL.String()
		seller := e.ChildText(".mbg-nw")
		fmt.Println(title, dateFinal, cost, shipping, seller)

		details := make(map[string]string, 0)

		keys := make([]string, 0)
		values := make([]string, 0)
		e.ForEach(".ux-layout-section__item > div", func(_ int, row *colly.HTMLElement) {
			row.ForEach(".ux-labels-values__labels-content", func(i int, el *colly.HTMLElement) {
				keys = append(keys, el.Text)
			})
			row.ForEach(".ux-labels-values__values-content", func(i int, el *colly.HTMLElement) {
				values = append(values, el.Text)
			})
		})

		for i, key := range keys {
			if strings.Index(key, "Artikelzustand") != -1 {
				continue
			}
			details[key] = values[i]
		}

		sale := sale.Sale{
			Title:            title,
			DateSold:         dateFinal,
			DateScraped:      time.Now(),
			Cost:             StringToCents(cost),
			ProposalAccepted: proposalAccepted,
			Shipping:         StringToCents(shipping),
			URL:              url,
			Seller:           seller,
			Detail:           details,
		}

		*sales = append(*sales, sale)

	})
	return c, nil
}

func StringToCents(input string) int {
	var euros, cents int
	if strings.Index(input, ",") == -1 {
		euros, _ = strconv.Atoi(input)
	} else {
		inputs := strings.Split(input, ",")
		cents, _ = strconv.Atoi(inputs[1])
		euros, _ = strconv.Atoi(inputs[0])

	}
	return euros*100 + cents
}
