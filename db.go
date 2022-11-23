package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func prepareDB() error {
	db, err := sql.Open("sqlite3", "./pollen.db")
	if err != nil {
		return err
	}
	defer db.Close()

	schema := `
create table if not exists pollen_sites (
    id integer primary key,
    name text
);

create table if not exists pollen (
    site integer,
    type text,
    severity text,
    timestamp date,
                                  
    foreign key (site)
    	references pollen_sites(id)
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

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert or ignore into pollen_sites(id, name) values(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for siteID, siteName := range Sites {
		if _, err = stmt.Exec(siteID, siteName); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func insertForecast(f Forecast) error {
	db, err := sql.Open("sqlite3", "./pollen.db")
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

func latestEntry() (t time.Time, err error) {
	db, err := sql.Open("sqlite3", "./pollen.db")
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
