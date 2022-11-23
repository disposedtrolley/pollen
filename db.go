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
    severity text,
    timestamp date
);

create table if not exists thunderstorm_asthma (
    region text,
    severity text,
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
	stmt, err := tx.Prepare("insert into pollen(site, type, severity, timestamp) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, pollenForecast := range f.Pollen {
		for _, severity := range pollenForecast.Severities {
			if _, err = stmt.Exec(pollenForecast.Site, severity.Type, severity.Severity, f.Date); err != nil {
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
	stmt, err = tx.Prepare("insert into thunderstorm_asthma(region, severity, timestamp) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, asthmaForecast := range f.ThunderstormAsthma {
		if _, err = stmt.Exec(asthmaForecast.Region, asthmaForecast.Severity, f.Date); err != nil {
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
		pollenForSite := Pollen{
			Site: siteName,
		}

		stmt, err := db.Prepare("select type, severity, timestamp from pollen where date(timestamp) == ? and site == ?")
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
			var pollenSeverity PollenSeverity
			err = rows.Scan(&pollenSeverity.Type, &pollenSeverity.Severity, &f.Date)
			if err != nil {
				return f, err
			}

			pollenForSite.Severities = append(pollenForSite.Severities, pollenSeverity)
		}
		err = rows.Err()
		if err != nil {
			return f, err
		}

		f.Pollen = append(f.Pollen, pollenForSite)
	}

	// Thunderstorm Asthma
	stmt, err := db.Prepare("select region, severity, timestamp from thunderstorm_asthma where date(timestamp) == ?")
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
		var asthma ThunderstormAsthma
		err = rows.Scan(&asthma.Region, &asthma.Severity, &f.Date)
		if err != nil {
			return f, err
		}

		f.ThunderstormAsthma = append(f.ThunderstormAsthma, asthma)
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
