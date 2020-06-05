.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-d -s -w" -tags netgo -installsuffix netgo -o bin/setWebhook setWebhook/main.go
	env GOOS=linux go build -ldflags="-d -s -w" -tags netgo -installsuffix netgo -o bin/webhook webhook/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
