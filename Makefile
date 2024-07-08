build: 
	@go build -o bin/rediscl .

run: build
	@./bin/rediscl 
test: 
	@go test ./... -v

docker_build:
	@echo building docker image with name rediscl
	docker build -t rediscl .
	@echo container is  ready to use

docker_run: docker_build
	@echo staring container is detach mod on port 6666
	docker run --name myredis -p 6666:6666 -d  rediscl
	@echo container is running

docker_rm: 
	@echo removing container and image 
	docker stop myredis && docker rm myredis
	docker rmi rediscl
	@echo done 