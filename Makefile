.PHONY: clean init run

all: clean init run

init:
	go mod tidy
	docker-compose up --build --no-start && docker-compose start

run:
	go run cmd/main.go

clean:
	docker-compose down --volumes