APP=workout-backend
APP_EXECUTABLE="./out/${APP}"

ALL_PACKAGES=$(shell go list ./... | grep -v "vendor" | grep -v "cmd/scripts/")

app:
	mkdir -p ./out
	go build -o ${APP_EXECUTABLE} ./cmd/app/... 

migrate:
	${APP_EXECUTABLE} migrate --config config.json

rollback:
	${APP_EXECUTABLE} rollback --config config.json

seed:
	${APP_EXECUTABLE} seed --config config.json

run_gql_playground:
	${APP_EXECUTABLE} gql-playground --config config.json

generate_gql_code:
	go run github.com/99designs/gqlgen generate

run_server:
	${APP_EXECUTABLE} server --config config.json
