include .env
export


service-run:
	 go run main.go

install-run:
	go get github.com/golang-migrate/migrate/v4 && migrate create -ext sql -dir migrations -seq init && go get github.com/joho/godotenv && \
	go get -u go.uber.org/zap

migrate-up:
	migrate -path migrations -database ${DATABASE_URL} up

migrate-down:
	migrate -path migrations -database ${DATABASE_URL} down

migrate-force:
	migrate -path migrations -database "postgres://postgres:dani123l@localhost:5432/library?sslmode=disable" force 1

docker-build:
	docker build . -t study-http-image

docker-entrance:
	docker run -d -p 5050:5050 study-http-image:latest

docker-deploy:
	docker compose up -d application

docker-nodeploy:
	docker compose down -d application

docker-volume:
	docker run --volume ./docker:/app/docker volumetest:latest

docker-postgres:
	docker run -e POSTGRES_PASSWORD=dani123l- -p 5432:5432 -v ./docker/pgdata:/var/lib/postgresql -d postgres:18-bookworm