build: 
	@go build -o bin/rediscl .

run: build
	@./bin/rediscl 
test: 
	@go test ./... -v
