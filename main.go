package main

import (
	
	"log"

	"github.com/gocolly/colly"

	"scrap/constant"
	api "scrap/googleDocsApi"
)

func main() {
	heading, scrapTable, err := Scraping(constant.URL)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	srv, err := api.GetService(constant.NameFileKeyGoogleApi)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	//Создание траблицы
	err = api.CreatTable(srv, heading, scrapTable)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	
	// Синхронизация таблицы

	// err = api.Updatetable(srv, heading, scrapTable)
	// if err != nil {
	// 	log.Fatal("Error: ", err)
	// }
}


// ------------------------- Получение таблицы по заданному URL -------------------------
func Scraping(url string) (constant.Heading, []constant.TableResponseCodes, error) {
	c := colly.NewCollector()
	var scrapTable []constant.TableResponseCodes
	var heading constant.Heading

	c.OnHTML("." + constant.NameClassTableHTML + " > tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			tableData := constant.TableResponseCodes{
				Code:        el.ChildText("td:nth-child(1)"),
				Desctiption: el.ChildText("td:nth-child(2)"),
			}
			scrapTable = append(scrapTable, tableData)
		})
	})

	c.OnHTML("." + constant.NameClassTableHTML + " > thead", func(e *colly.HTMLElement) {
		heading.Code = e.DOM.Find("th:nth-child(1)").Text()
		heading.Desctiption = e.DOM.Find("th:nth-child(2)").Text()
	})

	err := c.Visit(url)
	if err != nil {
		return constant.Heading{}, nil, err
	}

	return heading, scrapTable, nil
}