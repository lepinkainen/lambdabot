FUNCNAME=lambdabot
BINARYNAME=bootstrap
BUILDDIR=build

.PHONY: build

# from https://unix.stackexchange.com/a/235254
-include .env
export $(shell sed 's/=.*//' .env)


build: test
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(BUILDDIR)/$(BINARYNAME)
	cd $(BUILDDIR) && zip $(FUNCNAME).zip $(BINARYNAME)

test:
	go vet ./...
	go test -v ./...

lint:
	-golangci-lint run ./...

publish: test lint build
	aws lambda update-function-code --publish --function-name $(FUNCNAME) --zip-file fileb://$(BUILDDIR)/$(FUNCNAME).zip
