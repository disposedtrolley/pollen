package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var dsn string

func init() {
	dsn = "./pollen.db"

	if connStr, ok := os.LookupEnv("DATABASE_URL"); ok {
		dsn = connStr
	}
}

func prepareDB() error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	schema := `
create table if not exists pollen (
    site text,
    type text,
    forecast text,
    timestamp date
);

create table if not exists thunderstorm_asthma (
    region text,
    forecast text,
    timestamp date
);
`

	if _, err = db.Exec(schema); err != nil {
		return err
	}

	return nil
}

func insertForecast(f Forecast) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Pollen
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into pollen(site, type, forecast, timestamp) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, pollenSite := range f.Pollen.Sites {
		for _, prediction := range pollenSite.Predictions {
			if _, err = stmt.Exec(pollenSite.Site, prediction.Type, prediction.Severity, pollenSite.Date); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	// Thunderstorm Asthma
	tx, err = db.Begin()
	if err != nil {
		return err
	}
	stmt, err = tx.Prepare("insert into thunderstorm_asthma(region, forecast, timestamp) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, asthmaForecast := range f.ThunderstormAsthma.Predictions {
		if _, err = stmt.Exec(asthmaForecast.Region, asthmaForecast.Severity, f.ThunderstormAsthma.Date); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func selectForecast(date time.Time) (f Forecast, err error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return f, err
	}
	defer db.Close()

	y, m, d := date.Date()
	dateString := fmt.Sprintf("%d-%d-%d", y, m, d)

	// Pollen
	for _, siteName := range Sites {
		pollenForSite := PollenSite{
			Site: siteName,
		}

		stmt, err := db.Prepare("select type, forecast, timestamp from pollen where date(timestamp) == ? and site == ?")
		if err != nil {
			return f, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(dateString, siteName)
		if err != nil {
			return f, err
		}
		defer rows.Close()

		for rows.Next() {
			var pollenPrediction PollenPrediction
			err = rows.Scan(&pollenPrediction.Type, &pollenPrediction.Severity, &pollenForSite.Date)
			if err != nil {
				return f, err
			}

			pollenForSite.Predictions = append(pollenForSite.Predictions, pollenPrediction)
		}
		err = rows.Err()
		if err != nil {
			return f, err
		}

		f.Pollen.Sites = append(f.Pollen.Sites, pollenForSite)
	}

	// Thunderstorm Asthma
	stmt, err := db.Prepare("select region, forecast, timestamp from thunderstorm_asthma where date(timestamp) == ?")
	if err != nil {
		return f, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(dateString)
	if err != nil {
		return f, err
	}
	defer rows.Close()

	for rows.Next() {
		var asthma ThunderstormAsthmaPrediction
		err = rows.Scan(&asthma.Region, &asthma.Severity, &f.ThunderstormAsthma.Date)
		if err != nil {
			return f, err
		}

		f.ThunderstormAsthma.Predictions = append(f.ThunderstormAsthma.Predictions, asthma)
	}
	err = rows.Err()
	if err != nil {
		return f, err
	}

	return f, nil
}

func latestEntry() (t time.Time, err error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return t, err
	}
	defer db.Close()

	var timestamp time.Time
	row := db.QueryRow("select timestamp from pollen order by timestamp desc limit 1")

	if err := row.Scan(&timestamp); err != nil {
		return t, err
	}

	return timestamp, nil
}
