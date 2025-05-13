start:
	docker-compose up

stop:
	docker-compose down

restart:
	docker-compose down && docker-compose up

lint:
	golangci-lint run ./... --config=./.golangci.yml

test:
	go test ./... -cover

