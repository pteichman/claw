package claw

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Note struct {
	PK int `db:"Z_PK"`

	Title     sql.NullString `db:"ZTITLE"`
	Text      sql.NullString `db:"ZTEXT"`
	Archived  bool           `db:"ZARCHIVED"`
	Encrypted bool           `db:"ZENCRYPTED"`
	Trashed   bool           `db:"ZTRASHED"`
}

func (n Note) TitleFlags() string {
	buf := bytes.NewBufferString(n.Title.String)

	if n.Archived {
		fmt.Fprint(buf, " [archived]")
	}

	if n.Encrypted {
		fmt.Fprint(buf, " [encrypted]")
	}

	if n.Trashed {
		fmt.Fprint(buf, " [trashed]")
	}

	return buf.String()
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

type Link struct {
	ByNote int `db:"Z_7LINKEDBYNOTES"`
	Note   int `db:"Z_7LINKEDNOTES"`
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

func maxPK(notes []Note) int {
	var max int
	for _, n := range notes {
		if n.PK > max {
			max = n.PK
		}
	}
	return max
}
