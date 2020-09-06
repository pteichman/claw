package claw

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os/user"
	"path"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func cmdUnlinked(args []string, stdout, stderr io.Writer) int {
	var app unlinkedApp
	if err := app.fromArgs(args, stdout, stderr); err != nil {
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(stderr, "Runtime error: %v\n", err)
		return 1
	}

	return 0
}

type unlinkedApp struct {
	dbpath string

	logger *log.Logger
	stdout io.Writer
}

func (app *unlinkedApp) fromArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("unlinked", flag.ContinueOnError)
	flags.SetOutput(stderr)

	flags.StringVar(
		&app.dbpath, "dbpath", "", "Database file",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	app.logger = log.New(stderr, "", log.LstdFlags)
	app.stdout = stdout

	if app.dbpath == "" {
		u, err := user.Current()
		if err != nil {
			app.logger.Fatalf("You don't exist: %s", err)
		}
		app.dbpath = path.Join(u.HomeDir, "/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite")
	}

	return nil
}

func (app *unlinkedApp) run() error {
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

	// List the leaf notes: those that are not linked by anything else.
	found := make([]bool, maxPK(notes)+1)
	for _, link := range links {
		found[link.Note] = true
	}

	for _, note := range notes {
		if !found[note.PK] {
			fmt.Fprintln(app.stdout, note.TitleFlags())
		}
	}

	return nil
}
