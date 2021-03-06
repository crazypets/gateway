.PHONY: lint build swag_mix up down

lint:
	golangci-lint run --enable-all

build:
	go build -o auth cmd/gateway/main.go ;\

up:
	docker-compose -f deployments/docker-compose.yml up -d --build;\
	docker image prune -f ;\

down:
	docker-compose -f deployments/docker-compose.yml down ;\

swag_mix:
	docker pull quay.io/goswagger/swagger ;\
	alias swagger="docker run --rm -it -e GOPATH=$HOME/go:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger" ;\
	swagger mixin -o api/swagger/gateway.swagger.json \
		api/swagger/auth.swagger.json \
		api/swagger/pcd.swagger.json \
		api/swagger/file_storage.swagger.json \
		api/swagger/teo.swagger.json \;
