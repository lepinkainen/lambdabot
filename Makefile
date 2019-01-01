FUNCNAME=lambdabot
BUILDDIR=build

.PHONY: build

# from https://unix.stackexchange.com/a/235254
include .env
export $(shell sed 's/=.*//' .env)


build: test
	env GOOS=linux GOARCH=amd64 go build -o $(BUILDDIR)/$(FUNCNAME)
	cd $(BUILDDIR) && zip $(FUNCNAME).zip $(FUNCNAME)

test:
	go test -v ./...

publish: test build
	aws lambda update-function-code --publish --function-name $(FUNCNAME) --zip-file fileb://$(BUILDDIR)/$(FUNCNAME).zip
