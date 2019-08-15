package shortlinks

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

// setupServer starts a httptest.Server and returns the servers base URL
// as well as a function to stop the server after testing finished.
func setupServer() (base *url.URL, tearDown func()) {
	i, err := NewSQLiteIndex("file:shortlinks_test.sqlite?_journal_mode=wal")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s := Server{
		index: i,
		quiet: true,
	}

	r := chi.NewRouter()

	s.SetupRoutes(r)

	ts := httptest.NewServer(r)

	base, _ = url.Parse(ts.URL) // this should never fail

	return base, ts.Close
}

var linkToShortURL = regexp.MustCompile(`<a href="/([a-z0-9]+)">[^<]+</a>`)

func TestServer(t *testing.T) {
	type expectedResponse struct {
		shouldRedirect          bool
		longURLInLocationHeader string
		shortURLPath            string
	}

	shortenResponse := func(expectedShortURLPath string) expectedResponse {
		return expectedResponse{
			shortURLPath: expectedShortURLPath,
		}
	}

	redirectResponse := func(expectedLongURL string) expectedResponse {
		return expectedResponse{
			shouldRedirect: true, longURLInLocationHeader: expectedLongURL,
		}
	}

	type testRequest struct {
		path string
		expectedResponse
	}

	tests := []struct {
		name     string
		requests []testRequest
	}{
		{
			name: "basic",
			requests: []testRequest{
				{
					"/https://basic.com",
					shortenResponse("/1"),
				},
				{
					"/1",
					redirectResponse("https://basic.com/"),
				},
				{
					"/https://basic-2.com",
					shortenResponse("/2"),
				},
				{
					"/2",
					redirectResponse("https://basic-2.com/"),
				},
			},
		},
		{
			name: "IDs are persistent across sequential requests",
			requests: []testRequest{
				{
					"/https://ids.com",
					shortenResponse("/3"),
				},
				{
					"/https://ids.com",
					shortenResponse("/3"),
				},
				{
					"/3",
					redirectResponse("https://ids.com/"),
				},
			},
		},
		{
			name: "slash is used instead of empty path",
			requests: []testRequest{
				{
					"/https://slash.com",
					shortenResponse("/4"),
				},
				{
					"/4",
					redirectResponse("https://slash.com/"),
				},
				{
					"/https://slash.com/",
					shortenResponse("/4"),
				},
				{
					"/4",
					redirectResponse("https://slash.com/"),
				},
			},
		},
	}

	// make sure our HTTP client does not follow redirects, since we need to test for them
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, test := range tests {
		_base, tearDown := setupServer()

		for _, req := range test.requests {
			reqURL, _ := url.Parse(_base.String())
			reqURL.Path = req.path
			reqURL.RawPath = ""

			errorf := func(format string, args ...interface{}) {
				t.Errorf("request '%s' of test '%s' failed: "+format, append([]interface{}{reqURL.String(), test.name}, args...)...)
			}

			// we wrap the following code in a func() closure so we can make use of return and defer
			func() {
				resp, err := client.Get(reqURL.String())
				if err != nil {
					errorf("error during GET: %v", err)
					return
				}
				defer resp.Body.Close()                  // make sure we close the response body
				defer io.Copy(ioutil.Discard, resp.Body) // make sure we read the full response body (runs before the above deferred call)

				if req.shouldRedirect {
					if resp.StatusCode != http.StatusFound {
						errorf("expected redirect (%d), but got %d", http.StatusFound, resp.StatusCode)
						return
					}
					if resp.Header.Get("Location") != req.longURLInLocationHeader {
						errorf("expected Location header value '%s', but got '%s'", req.longURLInLocationHeader, resp.Header.Get("Location"))
						return
					}
				} else {
					_body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						errorf("error while reading response body: %v", err)
						return
					}
					body := string(_body)
					body = strings.TrimSpace(body)
					shortURL, err := url.Parse(body)
					if err != nil {
						errorf("error parsing response as URL: %v", err)
						return
					}

					if shortURL.Path != req.shortURLPath {
						errorf("expected path '%s' in short URL, but found %s", req.shortURLPath, shortURL.Path)
						return
					}
				}
			}()
		}

		tearDown()
	}
}
