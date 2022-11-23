package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

const (
	appDataUrl = "https://api.pollenforecast.com.au/app/json/app_data.php?app=1&version=4"
)

type Forecast struct {
	Date               time.Time
	ThunderstormAsthma ThunderstormAsthma
}

type ThunderstormAsthma struct {
	Region   string
	Severity string
}

type AppData struct {
	ThunderstormAsthma string `json:"div8"`
}

func getThunderstormAsthma() (forecast []ThunderstormAsthma, err error) {
	resp, err := http.Get(appDataUrl)
	if err != nil {
		return forecast, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return forecast, err
	}

	var appData AppData
	err = json.Unmarshal(body, &appData)
	if err != nil {
		return forecast, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(appData.ThunderstormAsthma))
	if err != nil {
		return forecast, err
	}

	tbody := htmlquery.FindOne(doc, "//tbody")
	rows, err := htmlquery.QueryAll(tbody, "//tr")
	if err != nil {
		return forecast, err
	}

	for _, row := range rows {
		forecastText := htmlquery.InnerText(row)
		forecastRow := strings.Split(forecastText, "\n")

		forecast = append(forecast, ThunderstormAsthma{
			Region:   strings.Trim(forecastRow[1], " "),
			Severity: strings.Trim(forecastRow[3], " "),
		})
	}

	return forecast, nil
}

func main() {
	thunderstormAsthma, err := getThunderstormAsthma()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(thunderstormAsthma)
}
