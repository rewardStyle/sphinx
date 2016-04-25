.PHONY: build
PKGS := $(shell glide nv)

build: fixture_generator
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build

fixture_generator:
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build -o fixture_data/generate_fixtures fixture_data/generate_fixtures.go

clean:
	rm fixture_data/generate_fixtures
	rm fixture_data/generated/*.tst

fixturedata: fixture_generator
	fixture_data/generate_fixtures

test: fixturedata
	go test -short $(PKGS)

testall: fixturedata
	go test $(PKGS)
