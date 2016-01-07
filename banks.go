package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	str "strings"
	"time"
)

var (
	kbz = "http://www.kbzbank.com"
	cb  = "http://www.cbbank.com.mm/exchange_rate.aspx"
)

func scrapKBZ(url string) ([]string, string) {
	tmp := []string{}

	doc, err := goquery.NewDocument(url)
	PanicIf(err)

	doc.Find(".answer tbody tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td").Each(func(u int, t *goquery.Selection) {
			tmp = append(tmp, str.TrimSpace(t.Text()))
		})
	})

	return tmp, "kbz"
}

func scrapAGD() ([]string, string) {
	tmp := []string{}
	// Using with file
	f, err := os.Open("agd.html")
	PanicIf(err)
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	doc.Find("#curency-table tbody tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td").Each(func(u int, t *goquery.Selection) {
			tmp = append(tmp, str.TrimSpace(t.Text()))
		})
	})

	return tmp, "agd"
}

func scrapCB(url string) ([]string, string) {
	tmp := []string{}

	doc, err := goquery.NewDocument(url)
	PanicIf(err)

	doc.Find("table tr").Slice(1, 4).Find("td").Each(func(i int, s *goquery.Selection) {
		tmp = append(tmp, str.TrimSpace(s.Text()))
	})

	return tmp, "cb"
}

func process(tmp []string, bName string) Bank {
	bank := Bank{}

	if bName == "kbz" {
		bank.Name = "KBZ"
	} else if bName == "cb" {
		bank.Name = "CB"
	} else if bName == "agd" {
		bank.Name = "AGD"
	}

	bank.Base = "MMK"
	bank.Time = time.Now().String()

	currencies := []string{tmp[0], tmp[3], tmp[6]}
	buy := []string{tmp[1], tmp[4], tmp[7]}
	sell := []string{tmp[2], tmp[5], tmp[8]}

	for x, _ := range currencies {
		bank.Rates = append(bank.Rates, map[string]BuySell{
			currencies[x]: BuySell{buy[x], sell[x]}})
	}

	return bank
}

func main() {

	rawAGD, agd := scrapAGD()
	rawKBZ, kbz := scrapKBZ(kbz)
	rawCB, cb := scrapCB(cb)

	r := gin.Default()

	r.GET("/:bank", func(c *gin.Context) {
		bankName := c.Params.ByName("bank")

		if bankName == "kbz" {
			bank := process(rawKBZ, kbz)
			c.JSON(200, bank)
		} else if bankName == "cb" {
			bank := process(rawCB, cb)
			c.JSON(200, bank)
		} else if bankName == "agd" {
			bank := process(rawAGD, agd)
			c.JSON(200, bank)
		}
	})

	r.GET("/", func(c *gin.Context) {
		body, err := ioutil.ReadFile("README.md")
		PanicIf(err)
		c.String(200,
			string(blackfriday.MarkdownBasic([]byte(body))))
	})

	r.Run(":" + os.Getenv("PORT"))
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

type Bank struct {
	Name  string               `json:"name"`
	Base  string               `json:"base"`
	Time  string               `json:"time"`
	Rates []map[string]BuySell `json:"rates"`
}

type BuySell struct {
	Buy  string `json:"buy"`
	Sell string `json:"sell"`
}