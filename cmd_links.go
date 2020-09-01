package claw

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os/user"
	"path"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func cmdLinks(args []string, stdout, stderr io.Writer) int {
	var app linksApp
	if err := app.fromArgs(args, stdout, stderr); err != nil {
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(stderr, "Runtime error: %v\n", err)
		return 1
	}

	return 0
}

type linksApp struct {
	dbpath string

	logger *log.Logger
}

func (app *linksApp) fromArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("links", flag.ContinueOnError)
	flags.SetOutput(stderr)

	flags.StringVar(
		&app.dbpath, "dbpath", "", "Database file",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	app.logger = log.New(stderr, "", log.LstdFlags)

	if app.dbpath == "" {
		u, err := user.Current()
		if err != nil {
			app.logger.Fatalf("You don't exist: %s", err)
		}
		app.dbpath = path.Join(u.HomeDir, "/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite")
	}

	return nil
}

func (app *linksApp) run() error {
	db, err := sqlx.Open("sqlite3", app.dbpath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Put the database in read only mode for extra safety.
	if _, err = db.Exec("PRAGMA query_only = true;"); err != nil {
		return err
	}

	notes, err := fetchNotes(db)
	if err != nil {
		return err
	}

	links, err := fetchLinks(db)
	if err != nil {
		return err
	}

	nmap := make(map[int]Note)
	for _, n := range notes {
		nmap[n.PK] = n
	}

	for _, link := range links {
		a := nmap[link.ByNote]
		b := nmap[link.Note]

		if b.Trashed {
			app.logger.Printf("[%d]%s -> [%d]%s [trash]", a.PK, a.Title.String, b.PK, b.Title.String)
		} else {
			app.logger.Printf("[%d]%s -> [%d]%s", a.PK, a.Title.String, b.PK, b.Title.String)
		}
	}

	return nil
}

type Note struct {
	PK int `db:"Z_PK"`

	Title     sql.NullString `db:"ZTITLE"`
	Text      sql.NullString `db:"ZTEXT"`
	Archived  bool           `db:"ZARCHIVED"`
	Encrypted bool           `db:"ZENCRYPTED"`
	Trashed   bool           `db:"ZTRASHED"`
}

func fetchNotes(db *sqlx.DB) ([]Note, error) {
	rows, err := db.Queryx("SELECT Z_PK, ZTITLE, ZTEXT, ZARCHIVED, ZENCRYPTED, ZTRASHED FROM ZSFNOTE")
	if err != nil {
		return nil, err
	}

	var notes []Note
	for rows.Next() {
		var note Note
		err = rows.StructScan(&note)
		if err != nil {
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func fetchLinks(db *sqlx.DB) ([]Link, error) {
	rows, err := db.Queryx("SELECT Z_7LINKEDBYNOTES, Z_7LINKEDNOTES FROM Z_7LINKEDNOTES ORDER BY Z_7LINKEDBYNOTES, Z_7LINKEDNOTES")
	if err != nil {
		return nil, err
	}

	var links []Link
	for rows.Next() {
		var link Link
		err = rows.StructScan(&link)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, nil
}

type Link struct {
	ByNote int `db:"Z_7LINKEDBYNOTES"`
	Note   int `db:"Z_7LINKEDNOTES"`
}
