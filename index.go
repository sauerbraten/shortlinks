package shortlinks

import (
	"context"
	"errors"
)

// Index keeps track of the ID -> URL mapping.
type Index interface {
	LookupID(id int64) (longURL string, err error)
	AddURL(longURL string, ctx context.Context) (id int64, err error)
}

var ErrNotFound = errors.New("not found in index")
