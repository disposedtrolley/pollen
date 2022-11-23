package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	appDataUrl         = "https://api.pollenforecast.com.au/app/json/app_data.php?app=1&version=4"
	dataAcquisitionUrl = "https://api.pollenforecast.com.au/app/json/data_acquisition.php?app=1&version=4"
)

var Sites = map[int]string{
	1:  "melbourne",
	5:  "dookie",
	6:  "bendigo",
	7:  "creswick",
	8:  "hamilton",
	9:  "churchill",
	15: "burwood",
	16: "geelong",
}

type PollenType string

const (
	PollenGrass          = "grass"
	PollenTreeCypress    = "cypress"
	PollenTreeMyrtle     = "myrtle"
	PollenTreeOlive      = "olive"
	PollenTreePlane      = "plane"
	PollenTreeAlternaria = "alternaria"
	PollenWeedPlantain   = "plantain"
)

type Severity string

const (
	SeverityLow      = "low"
	SeverityModerate = "moderate"
	SeverityHigh     = "high"
	SeverityExtreme  = "extreme"
)

type Forecast struct {
	Date               time.Time
	ThunderstormAsthma []ThunderstormAsthma
	Pollen             []Pollen
}

type ThunderstormAsthma struct {
	Region   string
	Severity string
}

type Pollen struct {
	Site       int
	Severities []PollenSeverity
}

type PollenSeverity struct {
	Type     PollenType
	Severity Severity
}

type AppData struct {
	ThunderstormAsthma string `json:"div8"`
}

type DataAcquisition struct {
	Result string `json:"result"`
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
			Region:   strings.ToLower(strings.Trim(forecastRow[1], " ")),
			Severity: strings.ToLower(strings.Trim(forecastRow[3], " ")),
		})
	}

	return forecast, nil
}

func getPollen(siteID int) (forecast Pollen, err error) {
	resp, err := http.Get(fmt.Sprintf("%s&site_id=%d", dataAcquisitionUrl, siteID))
	if err != nil {
		return forecast, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return forecast, err
	}

	var dataAcquisition DataAcquisition
	err = json.Unmarshal(body, &dataAcquisition)
	if err != nil {
		return forecast, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(dataAcquisition.Result))
	if err != nil {
		return forecast, err
	}

	// grass
	grassPollenDiv, err := htmlquery.Query(doc, "//div[@id='pollenCount']//div[4]")
	if err != nil {
		return forecast, err
	}
	forecast.Severities = append(forecast.Severities, PollenSeverity{
		Type:     PollenGrass,
		Severity: Severity(strings.ToLower(htmlquery.InnerText(grassPollenDiv))),
	})

	otherPollenForecastsDiv, err := htmlquery.Query(doc, "//div[@id='other_forecasts'][1]")
	if err != nil {
		return forecast, err
	}

	otherPollenForecasts, err := htmlquery.QueryAll(otherPollenForecastsDiv, "//li/div[@class='other_card_wrapper']/div[@class='card']")
	if err != nil {
		return forecast, err
	}

	for _, f := range otherPollenForecasts {
		pollenTypeAnchor, err := htmlquery.Query(f, "//a")
		if err != nil {
			return forecast, err
		}

		pollenType := strings.ToLower(htmlquery.InnerText(pollenTypeAnchor))

		pollenSeverityDiv, err := htmlquery.Query(f, "//div[2]")
		if err != nil {
			return forecast, err
		}

		pollenSeverity := strings.ToLower(htmlquery.InnerText(pollenSeverityDiv))

		forecast.Severities = append(forecast.Severities, PollenSeverity{
			Type:     PollenType(pollenType),
			Severity: Severity(pollenSeverity),
		})
	}

	return forecast, err
}

func getAllPollen() (forecast []Pollen, err error) {
	for siteID, siteName := range Sites {
		log.Printf("Getting pollen forecast for site %v (%s)...\n", siteID, siteName)

		f, err := getPollen(siteID)
		if err != nil {
			return forecast, err
		}
		f.Site = siteID
		forecast = append(forecast, f)

		log.Println("Done.")
		time.Sleep(1 * time.Second)
	}

	return forecast, nil
}

func getForecast() (forecast Forecast, err error) {
	log.Println("Getting thunderstorm asthma forecast...")
	thunderstormAsthma, err := getThunderstormAsthma()
	if err != nil {
		return forecast, err
	}
	log.Println("Done.")

	log.Println("Getting pollen forecasts...")
	pollen, err := getAllPollen()
	if err != nil {
		return forecast, err
	}
	log.Println("Done.")

	forecast = Forecast{
		Date:               time.Now().UTC(),
		ThunderstormAsthma: thunderstormAsthma,
		Pollen:             pollen,
	}

	return forecast, nil
}

func isToday(t time.Time) bool {
	tLocal := t.Local()
	tNow := time.Now().Local()

	y1, m1, d1 := tLocal.Date()
	y2, m2, d2 := tNow.Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}

func forecast() {
	if err := prepareDB(); err != nil {
		log.Fatal(err)
	}

	lastEntryDate, err := latestEntry()
	if err != nil {
		// Continue if the DB is empty.
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
	}

	if isToday(lastEntryDate) {
		log.Println("Already populated today's forecast. Stopping.")
		return
	}

	forecast, err := getForecast()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Inserting forecast data...")
	if err := insertForecast(forecast); err != nil {
		log.Fatal(err)
	}
	log.Println("Done.")
}
