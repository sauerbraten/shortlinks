package shortlinks

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"golang.org/x/xerrors"
)

type Server struct {
	index  Index
	logger *log.Logger
	quiet  bool
}

// NewServer returns a new Server using index i and logger l.
// If l is nil, a default like the log package's default logger
// will be used, which means logs will appear on os.Stderr.
func NewServer(i Index, l *log.Logger) *Server {
	if l == nil {
		l = log.New(os.Stderr, "", log.LstdFlags)
	}

	return &Server{
		index:  i,
		logger: l,
	}
}

// SetupRoutes registers the shorten and resolve handlers on the router.
func (s *Server) SetupRoutes(r chi.Router) {
	r.HandleFunc("/http://*", s.shorten)
	r.HandleFunc("/https://*", s.shorten)
	r.HandleFunc("/{id:[a-z0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		// we proxy the call to s.resolve through this "middleware" to resolve the id URL parameter
		s.resolve(w, r, chi.URLParam(r, "id"))
	})
}

// writeShortLink writes a short link as an absolute URL to the client.
func writeShortLink(w http.ResponseWriter, r *http.Request, id int64) {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	fmt.Fprintf(w, "%s://%s/%s\n", scheme, r.Host, formatID(id))
}

// writeError writes a printf-formatted response using the specified certain status code to the client.
func writeError(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format+"\n", args...)
}

// shorten handles requests with an absolute URL as request path. shorten adds the URL to the index,
// then responds to the client with an absolute short URL which the client can use instead in the future.
func (s *Server) shorten(w http.ResponseWriter, r *http.Request) {
	urlToShorten := r.URL.RequestURI() // r.URL.RequestURI() is e.g. /http://example.com/
	urlToShorten = urlToShorten[1:]    // cut off leading slash

	// parse as URL for validation, normalize, then convert back to string
	_longURL, err := url.Parse(urlToShorten)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not parse '%s' as URL: %v", urlToShorten, err)
		return
	}
	normalize(_longURL)
	longURL := _longURL.String()

	// make sure it's in the index
	shortID, err := s.index.AddURL(longURL, r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error adding URL to index: %v", err)
		return
	}

	if !s.quiet {
		log.Println("shortened", longURL, "to", shortID)
	}

	writeShortLink(w, r, shortID)
}

// resolve handles all requests where the request path could be a base36 encoded ID.
// resolve parses the shortID from the request path and uses it to look for a URL
// mapped to it in the index. If successful, the request is redirected to that URL.
func (s *Server) resolve(w http.ResponseWriter, r *http.Request, shortID string) {
	id, err := parseID(shortID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing ID as base36 integer: %v", err)
	}

	longURL, err := s.index.LookupID(id)
	if err != nil {
		if xerrors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "unknown link ID %s (parsed as %d)", shortID, id)
			return
		}
		writeError(w, http.StatusInternalServerError, "error shortening URL: %v", err)
		return
	}

	if !s.quiet {
		log.Println("resolved", shortID, "to", longURL)
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

// converts an int64 to a string using base 36
func formatID(rowID int64) string {
	return strconv.FormatInt(rowID, 36)
}

// converts a shortlink ID string in base 36 to an int64
func parseID(shortID string) (int64, error) {
	return strconv.ParseInt(shortID, 36, 64)
}

// normalize makes sure URLs with no path segment end with a slash
// for example: 'http://example.com' becomes 'http://example.com/'
func normalize(u *url.URL) {
	if u.Path == "" {
		u.Path = "/"
	}
}
