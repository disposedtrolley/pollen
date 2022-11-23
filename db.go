package main

import (
	"database/sql"
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
