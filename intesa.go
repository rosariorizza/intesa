package main

import (
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Permetti tutte le origini
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
	}))

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, getWords(true))

		/*
			var res string
			for _, s := range getWords(true) {
				//fmt.Println("%s")
				res += s + "\n"
			}
			return c.String(http.StatusOK, res)*/
	})

	// Start server
	e.Logger.Fatal(e.Start(":8000"))

}

func getWords(easy bool) [10]string {
	// Crea un nuovo colly collector
	c := colly.NewCollector()

	var words [10]string
	var i = -1

	c.OnHTML("div", func(e *colly.HTMLElement) {

		value := e.Attr("div")

		if i < 0 {
			i++
		} else if i < 10 {
			words[i] = e.Text[1:]
			//fmt.Printf("%s\n", e.Text[1:])
			i++
		}

		c.Visit(e.Request.AbsoluteURL(value))
	})

	if easy {
		c.Visit("https://www.parolecasuali.it/?fs=10&fs2=1&Submit=Nuova+parola")
	} else {
		c.Visit("https://www.parolecasuali.it/?fs=10&fs2=0&Submit=Nuova+parola")
	}

	return words
}
