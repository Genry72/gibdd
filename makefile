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
install: ##Создаем базовый образ
	docker build -f "./tmp/Base.Dockerfile" -t gibdd_base_image:v1 --build-arg USERID=$(shell id -u) --build-arg GROUPID=$(shell id -g) "." ##Собираем базовый образ
	docker build -f "./tmp/yandexDisk.Dockerfile" -t yandexdisk_image:v1 --build-arg USERID=$(shell id -u) --build-arg GROUPID=$(shell id -g) "." ##Собираем образ диска
	docker run -d --name yandexdisk --restart unless-stopped --mount type=bind,source=$(current_dir)/yadisk/,target=/home/node/yadisk yandexdisk_image:v1 ##Запускаем диск
	# GOOS=linux go build -o ./gibdd ./main.go ##Билдим
	docker build -f "./tmp/Dockerfile" -t gibdd_image:v1 --build-arg USERID=$(shell id -u) --build-arg GROUPID=$(shell id -g) "." ##Собираем исполняемый образ
	docker run -d --env-file ./tmp/env --name gibdd --restart unless-stopped -v $(current_dir)/yadisk:/home/node/app/yadisk:rw gibdd_image:v1
	echo ОК
update: ##Обновляем в базовом образе исходник
	docker build -f "./tmp/Dockerfile" -t gibdd_image:v1 --build-arg USERID=$(shell id -u) --build-arg GROUPID=$(shell id -g) "." ##Собираем исполняемый образ
	docker run -d --env-file ./tmp/env --name gibdd --restart unless-stopped -v $(current_dir)/yadisk:/home/node/app/yadisk:rw gibdd_image:v1
	echo ОК
yandex: ##Создаем базовый образ
	mkdir yandex-disk-config
	echo $(iidYandex) > yandex-disk-config/iid
	echo $(passwdYandex) > yandex-disk-config/passwd
	echo auth=\"/home/node/.config/yandex-disk/passwd\" > yandex-disk-config/config.cfg
	echo dir=\"/yadisk\" >> yandex-disk-config/config.cfg
	echo proxy=\"no\" >> yandex-disk-config/config.cfg
	docker build -f "Base.Dockerfile" -t gibdd_base_image:v1 --build-arg USERID=$(shell id -u) --build-arg GROUPID=$(shell id -g) "." ##Собираем базовый образ
	docker build -f "yandexDisk.Dockerfile" -t yandexdisk_image:v1 "." ##Собираем образ диска
	docker run -d --name yandexdisk --restart unless-stopped -v $(current_dir)/yadisk/:/yadisk:z yandexdisk_image:v1 ##Запускаем диск
	echo ОК