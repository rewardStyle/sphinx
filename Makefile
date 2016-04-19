.PHONY: build
PKGS := $(shell glide nv)

build:
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build

clean:
	rm fixture_data/generate_fixtures
	rm fixture_data/generated/*.tst

fixturedata:
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build -o fixture_data/generate_fixtures fixture_data/generate_fixtures.go
	fixture_data/generate_fixtures

test: fixturedata
	go test 
