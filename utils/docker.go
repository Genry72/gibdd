package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func Docker(command, test string) {
	var localCommands []string
	var remoteCommands []string
	if os.Getenv("iidYandex") == "" || os.Getenv("passwdYandex") == "" {
		errorLog.Fatal("Не заданы переменные iidYandex либо passwdYandex")
	}
	if os.Getenv("remoteHost") == "" || os.Getenv("remoteUser") == "" {
		errorLog.Fatal("Не заданы переменные подключения к удаленному хосту: remoteHost либо remoteUser")
	}
	if test == "false" {
		infoLog.Println("Выполнение команд на удаленном сервере")
	} else {
		infoLog.Println("Выполнение команд на тестовом сервере")
	}
	if command == "install" {
		//Локально собираем конфиг для диска
		localCommands = []string{
			// Собираем конфиг для диска
			"mkdir ./tmp",
			"mkdir ./tmp/yandex-disk-config", //Создаем папку с конфигом для подключения к диску
			"echo $iidYandex > ./tmp/yandex-disk-config/iid",
			"echo $passwdYandex > ./tmp/yandex-disk-config/passwd",
			"echo auth=\"/home/node/.config/yandex-disk/passwd\" > ./tmp/yandex-disk-config/config.cfg",
			"echo dir=\"/home/node/yadisk\" >> ./tmp/yandex-disk-config/config.cfg",
			"echo proxy=\"no\" >> ./tmp/yandex-disk-config/config.cfg",
			"GOOS=linux go build -o ./tmp/gibdd ./main.go", //Билдим бинарник
			"tar -czf ./tmp/install.tar.gz ./tmp/gibdd ./env ./makefile ./yandexDisk.Dockerfile ./Dockerfile ./Base.Dockerfile ./tmp/yandex-disk-config",
		}
		localCmd(localCommands)
		//Отправляем файл на хост
		infoLog.Println("Собрали локальный архив")
		infoLog.Println("Отправляем архив на сервер")
		err := CopyFileToHost("./tmp/install.tar.gz", "./install.tar.gz", os.Getenv("remoteUser"), "./id_rsa", os.Getenv("remoteHost"), test)
		if err != nil {
			log.Println(err)
			return
		}
		infoLog.Println("Отправили архив на сервер")
		//Удаляем временную папку
		localCmd([]string{"rm -rf ./tmp"})
		remoteCommands = []string{
			"mkdir -p ./gibdd/tmp",                    //Рабочая папка для запуска контейнеа
			"tar -xf ./install.tar.gz -C ./gibdd/tmp", //распаковываем архив
			"rm -f ./install.tar.gz",
			"cd ./gibdd",
			"mkdir -p ./yadisk/sync/gibddBot/", //Создаем папку для диска, она же и для БД
			// "echo test | sudo -S -s",
			// "sudo -s",
			"docker rm --force -v yandexdisk",                                        //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep yandexdisk_image | awk '{print $3}')", //Удаляем изображение
			"docker rm --force -v gibdd",                                             //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')",      //Удаляем изображение
			"docker rmi $(docker images | grep gibdd_base_image | awk '{print $3}')", //Удаляем ,базовое изображение
			"make -f ./tmp/makefile install",                                         //Создаем базовый образ
			"rm -rf ./tmp",
			"exit",
			// "exit",
		}
		err = SshExec(os.Getenv("remoteHost"), "./id_rsa", os.Getenv("remoteUser"), remoteCommands, test)
		if err != nil {
			errorLog.Println(err)
		}
		infoLog.Println("Выполнили команды на удаленном хосте")
	}
	if command == "update" {
		//Локально собираем конфиг для диска
		localCommands = []string{
			// Компилируем
			"mkdir ./tmp",
			"GOOS=linux go build -o ./tmp/gibdd ./main.go", //Билдим бинарник
			"tar -czf ./tmp/install.tar.gz ./tmp/gibdd ./env ./makefile ./Dockerfile",
		}
		if runtime.GOOS == "darwin" { //Если это мак, то компилим в контейнере
			infoLog.Println("Компилируем исходник в контейнере")
			localCommands = []string{
				"mkdir ./tmp", //временная папка для бинарника
				"make build",
				"docker rmi $(docker images | grep build | awk '{print $3}')",
				"tar -czf ./tmp/install.tar.gz ./tmp/gibdd ./env ./makefile ./Dockerfile",
			}
		}
		localCmd(localCommands)
		//Отправляем файл на хост
		infoLog.Println("Собрали локальный архив")

		err := CopyFileToHost("./tmp/install.tar.gz", "./install.tar.gz", os.Getenv("remoteUser"), "./id_rsa", os.Getenv("remoteHost"), test)
		if err != nil {
			errorLog.Println(err)
			return
		}
		infoLog.Println("Отправили архив на сервер")
		//Удаляем временную папку
		localCmd([]string{"rm -rf ./tmp"})
		remoteCommands = []string{
			"mkdir -p ./gibdd/tmp",                    //Рабочая папка для запуска контейнеа
			"tar -xf ./install.tar.gz -C ./gibdd/tmp", //распаковываем архив
			"rm -f ./install.tar.gz",
			"cd ./gibdd",
			// "echo test | sudo -S -s",
			// "sudo -s",
			"docker rm --force -v gibdd",                                        //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')", //Удаляем изображение
			"make -f ./tmp/makefile update",                                     //Создаем базовый образ
			"rm -rf ./tmp",
			"exit",
			// "exit",
		}
		err = SshExec(os.Getenv("remoteHost"), "./id_rsa", os.Getenv("remoteUser"), remoteCommands, test)
		if err != nil {
			errorLog.Println(err)
		}
		infoLog.Println("Выполнили команды на удаленном хосте")
	}
	if command == "yandex" {
		localCommands = []string{
			"rm -rf ./yadisk",
			"docker rm --force -v yandexdisk", //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep yandexdisk_image | awk '{print $3}')", //Удаляем изображение
			// "docker rmi $(docker images | grep gibdd_base_image | awk '{print $3}')", //Удаляем ,базовое изображение
			"make yandex",
			"rm -rf ./tmp",
		}
		localCmd(localCommands)
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
			errorLog.Println(err)
		}
	}
}
