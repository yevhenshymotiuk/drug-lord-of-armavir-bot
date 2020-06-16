GOCMD=go
GOBUILD=GOOS=linux $(GOCMD) build -ldflags="-d -s -w" -tags netgo -installsuffix netgo

.PHONY: build clean deploy

build:
	$(GOBUILD) -o bin/setwebhook src/setwebhook/main.go
	$(GOBUILD) -o bin/webhook src/webhook/main.go

clean:
	$(GOCMD) clean
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
