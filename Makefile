GOPATH=`go env GOPATH`/bin
export PATH := /usr/local/go/bin:$(HOME)/go/bin:/usr/local/bin:$(PATH)

test: 
	scripts/test_coverage.sh

install-govulncheck: 
	go install golang.org/x/vuln/cmd/govulncheck@latest

vuln: install-govulncheck
	govulncheck ./...

install-golang-ci: 
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

ci-lint: install-golang-ci
	golangci-lint run ./...

make-lint: 
	make -n all 1>/dev/null

lint: make-lint go-lint 

go-lint: ci-lint vuln

all: lint test