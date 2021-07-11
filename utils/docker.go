package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Docker(command string) {
	var commands []string
	if command == "install" {
		commands = []string{ //Компилируем исходник внутри контейнера. Исполняемый файл запускаем в другом контейнере.
			"mkdir ./db",
			"docker rm --force -v gibdd", //Удаляем контейнер и его вольюм
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')",      //Удаляем изображение
			"docker rmi $(docker images | grep gibdd_base_image | awk '{print $3}')", //Удаляем ,базовое изображение
			"make install", //Создаем базовый образ
		}
	}
	if command == "update" {
		commands = []string{ //Компилируем исходник внутри контейнера. Исполняемый файл запускаем в другом контейнере.
			"mkdir ./db",
			"docker rm --force -v gibdd", //Удаляем контейнер
			"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')", //Удаляем изображение
			"make update", //обновляем бинарник в базовом образе
			// "docker system prune -a -f",
		}
	}
	if command == "yandexDisk" {
		commands = []string{ //Компилируем исходник внутри контейнера. Исполняемый файл запускаем в другом контейнере.
			"docker rm --force -v yandexdisk",                                        //Удаляем контейнер
			"docker rmi $(docker images | grep yandexdisk_image | awk '{print $3}')", //Удаляем изображение
			"make yandexDisk", //обновляем бинарник в базовом образе
			// "docker system prune -a -f",
		}
	}

	// Собираем архив
	for _, command := range commands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			err := fmt.Errorf("команда: %v Ошибка: %v", command, err)
			log.Println(err)
		}
	}
}
