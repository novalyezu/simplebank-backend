include .env

PG_DOCKER_NAME=postgres16sb
PG_HOST=${POSTGRES_HOST}
PG_PORT=${POSTGRES_PORT}
PG_USER=${POSTGRES_USER}
PG_PASS=${POSTGRES_PASSWORD}
PG_DATABASE=${POSTGRES_DATABASE}

postgrecreate:
	docker run --name ${PG_DOCKER_NAME} -p ${PG_PORT}:5432 -e POSTGRES_USER=${PG_USER} -e POSTGRES_PASSWORD=${PG_PASS} -d postgres:16-alpine

postgrestart:
	docker start ${PG_DOCKER_NAME}

postgrestop:
	docker stop ${PG_DOCKER_NAME}

createdb:
	docker exec -it ${PG_DOCKER_NAME} createdb --username=${PG_USER} --owner=${PG_USER} ${PG_DATABASE}

dropdb:
	docker exec -it ${PG_DOCKER_NAME} dropdb ${PG_DATABASE}

migrateup:
	migrate -path db/migration -database "postgresql://${PG_USER}:${PG_PASS}@${PG_HOST}:${PG_PORT}/${PG_DATABASE}?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://${PG_USER}:${PG_PASS}@${PG_HOST}:${PG_PORT}/${PG_DATABASE}?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: postgrecreate postgrestart postgrestop createdb dropdb migrateup migratedown sqlc
