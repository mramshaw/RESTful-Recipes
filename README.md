# RESTful recipes

[![Build status](https://travis-ci.org/mramshaw/RESTful-Recipes.svg?branch=master)](https://travis-ci.org/mramshaw/RESTful-Recipes)
[![Coverage Status](http://codecov.io/github/mramshaw/RESTful-Recipes/coverage.svg?branch=master)](http://codecov.io/github/mramshaw/RESTful-Recipes?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mramshaw/RESTful-Recipes?style=flat-square)](https://goreportcard.com/report/github.com/mramshaw/RESTful-Recipes)
[![GitHub release](https://img.shields.io/github/release/mramshaw/RESTful-Recipes.svg?style=flat-square)](https://github.com/mramshaw/RESTful-Recipes/releases)

A more formal REST API in Golang.

This builds on my [Simple REST API in Golang](https://github.com/mramshaw/Simple-REST-API).

All data is stored in [PostgreSQL](https://www.postgresql.org/), all transfer is via JSON.

All dependencies are handled via [Docker](https://www.docker.com/products/docker) and [docker-compose](https://github.com/docker/compose).

TDD (Test-Driven Development) is implemented; the build will fail if the tests do not pass.

Likewise the build will fail if either __golint__ or __go vet__ fails.

Enforces industry-standard __gofmt__ code formatting.

All testing can be done with [curl](CURLs.txt).


## Features

- uses [httprouter](https://github.com/julienschmidt/httprouter)
- uses [Pure Go postgres driver](https://github.com/lib/pq)
- uses [sqlx](#sqlx)
- uses [testify assertions](#testify-assertions)

#### sqlx

[sqlx](https://github.com/jmoiron/sqlx) offers some [database/sql](https://godoc.org/database/sql)
extensions.

There is a useful [Illustrated guide to SQLX](http://jmoiron.github.io/sqlx/).

It is not exactly a ___drop-in___ replacement for __database/sql__, as it is not quite a full
superset. I found that I still needed to import both dependencies due to some primitives (such
as __ErrNoRows__) that are not included in __sqlx__.

I am not entirely sure I would use this library in my normal workflow, although it does offer
some interesting new options (the __Illustrated Guide__ is very helpful here).

#### testify assertions

In terms of simplifying tests, [testify assertions](https://github.com/stretchr/testify/assert)
really make test cases much clearer - while also cutting down on boiler-plate.

I'm generally pretty reluctant to introduce additional external dependencies unless they really
deliver value. In my opinion this addition is justified as the test cases become much easier
to follow with assertions.


## To Run

The command to run:

    $ docker-compose up -d

For the first run, there will be a warning as `mramshaw4docs/golang-alpine:1.8` must be built.

This image will contain all of the Go dependencies and should only need to be built once.

For the very first run, `golang` may fail as it takes `postgres` some time to ramp up.

A successful `golang` startup should show the following as the last line of <kbd>docker-compose logs golang</kbd>

    golang_1    | 2018/02/24 18:38:01 Now serving recipes ...

If this line does not appear, repeat the `docker-compose up -d` command (there is no penalty for this).


## To Build:

The command to run:

    $ docker-compose up -d

Once `make` indicates that `restful_recipes` has been built, can change `docker-compose.yml` as follows:

1) Comment `command: make`

2) Uncomment `command: ./restful_recipes`


## For testing:

[Optional] Start postgres:

    $ docker-compose up -d postgres

Start golang [if postgres is not running, this step will start it]:

    $ docker-compose run --publish 80:8080 golang make

Successful startup will be indicated as follows:

    2018/02/24 16:27:10 Now serving recipes ...

This should make the web service available at:

    http://localhost/v1/recipes

Once the service is running, it is possible to `curl` it. Check `CURLs.txt` for examples.


## See what's running:

The command to run:

    $ docker ps


## View the build and/or execution logs

The command to run:

    $ docker-compose logs golang


## To Shutdown:

The command to run:

    $ docker-compose down


## Clean Up

The command to run:

    $ docker-compose run golang make clean


## To Do

- [x] Upgrade to latest Go (as of posting, 1.14.3)
- [x] Add Basic Auth to certain endpoints (POST, PUT/PATCH, DELETE)
- [x] 12-Factor Basic Auth parameters
- [x] Fix code coverage testing
- [ ] Upgrade to latest Postgres
- [ ] Persist back-end Postgres
- [ ] Add a SWAGGER definition
- [ ] Refactor data access into a DAO module
- [ ] Add tests for the DAO
- [ ] Add a health check
- [x] Migrate from Gorilla/mux to julienschmidt/httprouter
- [x] Migrate to [sqlx](https://github.com/jmoiron/sqlx) for some [sql](https://godoc.org/database/sql) extensions
- [x] Implement [testify](https://github.com/stretchr/testify) [assertions](https://godoc.org/github.com/stretchr/testify/assert)
- [ ] Implement [testify](https://github.com/stretchr/testify) [suite](https://godoc.org/github.com/stretchr/testify/suite)
- [ ] Implement CORS
- [ ] Implement graceful shutdown (available since __Go 1.8__)
- [ ] Add Prometheus-style instrumentation


## Credits

Inspired by this excellent tutorial by Kulshekhar Kabra:

    https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
