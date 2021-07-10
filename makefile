.DEFAULT_GOAL := help
current_dir=$(shell pwd)
ssh_prv_key=$(shell cat ~/.ssh/id_rsa)
ssh_pub_key=$(shell cat ~/.ssh/id_rsa.pub)
# Выводит описание целей - все, что написано после двойного диеза (##) через пробел
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'
build:
	docker-compose -f docker-compose.yml build
up: ## Создание и запуск контейнера
	# ## Удаляем бинарник
	# rm -rf ./app/test
	# docker build -f "Dockerfile" -t build:latest "." ##Билдим образ
	# ## Компилируем в папку тест
	# docker run -d --rm --mount type=volume,dst=/app,volume-driver=local,volume-opt=type=none,volume-opt=o=bind,volume-opt=device=$(current_dir) build:latest
	# sleep 10s; docker build -f "./app/Dockerfile" -t app:latest "." ##Билдим образ
	# docker rm -f $(shell docker ps -a -q  --filter ancestor=app:latest)
	# docker run -d -p 8080:8080 -d --restart unless-stopped app:latest
	# echo OK
	# Стартуем коуч с пробросом данных
	# docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 --mount type=volume,dst=/opt/couchbase/var,volume-driver=local,volume-opt=type=none,volume-opt=o=bind,volume-opt=device=$(current_dir)/var couchbase
	docker build -f "Dockerfile" -t health_img:latest "." ##Билдим образ
	# docker rm -f $(shell docker ps -a -q  --filter ancestor=health_img:latest)
	docker run -d --name healthCheck -p 2020:2020 --restart unless-stopped health_img:latest
	docker container prune -f
	echo ОК
base: ## Создаем базовый образ на клоне
	GOOS=linux go build -o ./gibdd ./main.go
	docker build -f "Dockerfile" -t gibdd_image:v1 "." ##Билдим и запускаем бинарник
	docker run -d --env-file ./env --name gibdd --restart unless-stopped --mount type=volume,dst=/app/db,volume-driver=local,volume-opt=type=none,volume-opt=o=bind,volume-opt=device=$(current_dir)/db gibdd_image:v1
	echo ОК
binaryClone: ##Для билда и запуска бинарника
	docker build -f "clone.Dockerfile" -t health:latest "." ##Собираем healthcheck
	docker run -d --env-file ./envClone --name healthCheck -p 2020:2020 --restart unless-stopped --hostname $(shell hostname) health:latest ##Запускаем healthcheck
	echo ОК
start:
	docker build .
	docker run test_build:latest --rm -v .:/app
down:
	docker-compose -f docker-compose.yml down $(c)
destroy:
	docker-compose -f docker-compose.yml down -v $(c)
stop:
	docker-compose -f docker-compose.yml stop $(c)
restart:
	docker-compose -f docker-compose.yml stop $(c)
	docker-compose -f docker-compose.yml up -d $(c)
logs:
	docker-compose -f docker-compose.yml logs --tail=100 -f $(c)
logs-api:
	docker-compose -f docker-compose.yml logs --tail=100 -f api
ps:
	docker-compose -f docker-compose.yml ps
login-timescale:
	docker-compose -f docker-compose.yml exec timescale /bin/bash
login-api:
	docker-compose -f docker-compose.yml exec api /bin/bash
db-shell:
	docker-compose -f docker-compose.yml exec timescale psql -Upostgres