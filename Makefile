FUNCNAME=lambdabot
BUILDDIR=build

.PHONY: build

build: test
	env GOOS=linux GOARCH=amd64 go build -o $(BUILDDIR)/$(FUNCNAME)
	cd $(BUILDDIR) && zip $(FUNCNAME).zip $(FUNCNAME)

test:
	go test -v ./...

publish: test build
	aws lambda update-function-code --function-name $(FUNCNAME) --zip-file fileb://$(BUILDDIR)/$(FUNCNAME).zip
