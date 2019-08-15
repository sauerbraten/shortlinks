package shortlinks

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/xerrors"
)

// SQLiteIndex is an Index backed by a SQLite database.
type SQLiteIndex struct {
	db *sql.DB
	l  *sync.Mutex // to synchronize DB access from concurrent requests
}

// compile-time assertion that we implement Index
var _ Index = &SQLiteIndex{}

// NewSQLiteIndex returns an Index backed by a SQLite database.
func NewSQLiteIndex(dsn string) (*SQLiteIndex, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, xerrors.Errorf("could not open SQLite database: %w", err)
	}

	return &SQLiteIndex{
		db: db,
		l:  new(sync.Mutex),
	}, nil
}

// LookupID returns the URL mapped to the provided ID.
func (i *SQLiteIndex) LookupID(id int64) (longURL string, err error) {
	i.l.Lock()
	defer i.l.Unlock()

	err = i.db.QueryRow("select longURL from links where id = ?", id).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", xerrors.Errorf("error resolving ID %s to long URL in database: %w", id, err)
	}

	return longURL, nil
}

// Add makes sure that a URL exists in the database by inserting it if neccessary, then
// returns the shortID belonging to that URL.
func (i *SQLiteIndex) AddURL(longURL string, ctx context.Context) (id int64, err error) {
	i.l.Lock()
	defer i.l.Unlock()

	_, err = i.db.Exec("insert or ignore into links (longURL) values (?)", longURL)
	if err != nil {
		return -1, xerrors.Errorf("error adding shortlink to database: %w", err)
	}

	err = i.db.QueryRow("select id from links where longURL = ?", longURL).Scan(&id)
	if err != nil {
		return -1, xerrors.Errorf("error getting ID of added URL from database: %w", err)
	}

	return id, nil
}
