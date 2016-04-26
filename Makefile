.PHONY: build
PKGS := $(shell glide nv)

build: fixture_generator
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build

fixture_generator:
	GO15VENDOREXPERIMENT=1 CGOENABLED=0 go build -o fixture_data/generate_fixtures fixture_data/generate_fixtures.go

clean:
	rm -f fixture_data/generate_fixtures
	rm -f fixture_data/generated/*.tst
	rm -f fixture_data/generated_header/*.tst

fixturedata: fixture_generator
	fixture_data/generate_fixtures

test: fixturedata
	go test -short $(PKGS)

testall: clean fixturedata
	go test $(PKGS)
