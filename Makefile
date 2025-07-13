.PHONY: build build-race up down start stop restart watch logs login

build:
	cd docker && docker compose build
build-race:
	cd docker && BUILD_WITH_RACE_DETECTION="1" docker compose build
up:
	cd docker && docker compose up -d
down:
	cd docker && docker compose down
start:
	cd docker && docker compose start
stop:
	cd docker && docker compose stop
restart: down up
watch:
	cd docker && export WATCH_FILES=1 && docker compose up -d
logs:
	cd docker && docker compose logs --tail=10 -f
login:
	docker exec -it api-golang-base-server sh
