#!/bin/sh
#
# Code coverage generation

export POSTGRES_HOST=${POSTGRES_HOST:=localhost}
export POSTGRES_PORT=${POSTGRES_PORT:=5432}
export POSTGRES_USER=${POSTGRES_USER:=epulze}
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:=epulze}
export POSTGRES_DB=${POSTGRES_DB:=tournaments}

COVERAGE_DIR="${COVERAGE_DIR:-coverage}"
PKG_LIST=$(go list ./... | grep -v /vendor/)

go test -i ./...

# Create the coverage files directory
mkdir -p "$COVERAGE_DIR";

# Create a coverage file for each package
for package in ${PKG_LIST}; do
    go test -covermode=count -coverprofile "${COVERAGE_DIR}/${package##*/}.cov" "$package" ;
done ;

# Merge the coverage profile files
echo 'mode: count' > "${COVERAGE_DIR}"/coverage.cov ;
tail -q -n +2 "${COVERAGE_DIR}"/*.cov >> "${COVERAGE_DIR}"/coverage.cov ;

# Display the global code coverage
go tool cover -func="${COVERAGE_DIR}"/coverage.cov ;

# Remove the coverage files directory
rm -rf "$COVERAGE_DIR";
