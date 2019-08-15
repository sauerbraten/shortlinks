# shortlinks

shortlinks is an easy to use link shortening service.

    $ make build
    $ shortlinks

`make build` (re)builds the shortlinks.sqlite database file and builds the `shortlinks` binary.

Shorten any URL by just using it as request path:

    $ curl -i http://localhost:5656/http://example.com/
    HTTP/1.1 200 OK
    Date: Mon, 12 Aug 2019 12:35:25 GMT
    Content-Length: 24
    Content-Type: text/plain; charset=utf-8

    http://localhost:5656/1

Resolve a shortlink to the long URL:

    $ curl -i http://localhost:5656/1
    HTTP/1.1 302 Found
    Content-Type: text/html; charset=utf-8
    Location: http://example.com/
    Date: Mon, 12 Aug 2019 12:36:00 GMT
    Content-Length: 41

    <a href="http://example.com/">Found</a>.

Most clients (including browsers) will automatically follow the redirect using the `Location: http://example.com/` header.

## More info

- IDs are base36 encoded for extra short links (base64 would be shorter, but conversion to and from int to base64 strings is not as convenient)
- ID->URL mapping is persisted to ensure it's kept between instance restarts
- uses a simple SQLite database (in write-ahead logging mode) as persistency layer, providing concurrent access for multiple service instances
- index interface to enable other persistency layers

## Testing

There are a few basic tests, ensuring correct behavior between request and response. To run them, use `make test`, which takes care of setting up and removing a test database.