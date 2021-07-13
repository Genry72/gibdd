package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Docker(command string) {
	var localCommands []string
	var remoteCommands []string
	if os.Getenv("iidYandex") == "" || os.Getenv("passwdYandex") == "" {
		log.Fatal("Не заданы переменные iidYandex либо passwdYandex")
	}
	if os.Getenv("remoteHost") == "" || os.Getenv("remoteUser") == "" {
		log.Fatal("Не заданы переменные подключения к удаленному хосту: remoteHost либо remoteUser")
	}
	if command == "install" {
		//Локально собираем конфиг для диска
		localCommands = []string{
			// Собираем конфиг для диска
			"rm -f ./install.tar.gz",
			"rm -rf yandex-disk-config",
			"mkdir yandex-disk-config", //Создаем папку с конфигом для подключения к диску
			"echo $iidYandex > ./yandex-disk-config/iid",
			"echo $passwdYandex > ./yandex-disk-config/passwd",
			"echo auth=\"/home/node/.config/yandex-disk/passwd\" > ./yandex-disk-config/config.cfg",
			"echo dir=\"/yadisk\" >> ./yandex-disk-config/config.cfg",
			"echo proxy=\"no\" >> ./yandex-disk-config/config.cfg",
			"GOOS=linux go build -o ./gibdd ./main.go", //Билдим бинарник
			"tar -czf ./install.tar.gz ./gibdd ./env ./makefile ./yandexDisk.Dockerfile ./Dockerfile ./Base.Dockerfile ./yandex-disk-config",
		}
		localCmd(localCommands)
		//Отправляем файл на хост
		log.Println("Собрали локальный архив")

		err := CopyFileToHost("./install.tar.gz", "./install.tar.gz", os.Getenv("remoteUser"), "./id_rsa", os.Getenv("remoteHost"))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Отправили архив на сервер")
		remoteCommands = []string{
			"rm -rf ./gibdd",
			"mkdir ./gibdd",                       //Рабочая папка для запуска контейнеа
			"tar -xf ./install.tar.gz -C ./gibdd", //распаковываем архив
			"cd ./gibdd",
			"mkdir -p ./yadisk/sync/gibddBot/", //Создаем папку для диска, она же и для БД
			"chmod -R 777 ./yadisk",
			"docker rm --force -v yandexdisk", //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep yandexdisk_image | awk '{print $3}')", //Удаляем изображение
			"docker rm --force -v gibdd",                                             //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')",      //Удаляем изображение
			"docker rmi $(docker images | grep gibdd_base_image | awk '{print $3}')", //Удаляем ,базовое изображение
			"make install",                //Создаем базовый образ
			"rm -rf ./yandex-disk-config", //Удаляем конфиг диска
			"rm -f ./gibdd",               //Удаляем исходник
			"exit",
		}
		err = SshExec(os.Getenv("remoteHost"), "./id_rsa", os.Getenv("remoteUser"), remoteCommands)
		if err != nil {
			log.Println(err)
		}
		log.Println("Выполнили команды на удаленном хосте")
	}
	if command == "update" {
		localCommands = []string{ //Компилируем исходник внутри контейнера. Исполняемый файл запускаем в другом контейнере.
			"docker rm --force -v gibdd",                                        //Удаляем контейнер
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')", //Удаляем изображение
			"make update", //обновляем бинарник в базовом образе
			// "docker system prune -a -f",
			"rm -f ./gibdd", //Удаляем исходник
		}
	}
}

//localCmd выполнение команд на локальном хосте
func localCmd(localCommands []string) {
	for _, command := range localCommands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			err := fmt.Errorf("команда: %v Ошибка: %v", command, err)
			log.Println(err)
		}
	}
}
