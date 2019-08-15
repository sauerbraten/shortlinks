.PHONY: all build rebuild_db server test test_clean clean

all: build test test_clean clean

build: rebuild_db server

rebuild_db:
	if [ -f shortlinks.sqlite ]; then rm shortlinks.sqlite*; fi
	sqlite3 shortlinks.sqlite < schema.sql

server:
	go build ./cmd/shortlinks/

test:
	if [ -f shortlinks_test.sqlite ]; then rm shortlinks_test.sqlite*; fi
	sqlite3 shortlinks_test.sqlite < schema.sql
	go test

test_clean:
	rm -f ./shortlinks_test.sqlite*

clean:
	rm -f ./shortlinks ./shortlinks.sqlite* ./shortlinks_test.sqlite*