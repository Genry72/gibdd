package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Docker() {
	commands := []string{//Пока локально. Собираем исходник внутри контейнера. Исполняемый файл запускаем в другом контейнере.
		"docker rm --force gibdd", //Удаляем контейнер
		"docker rmi $(docker images | grep gibdd_image | awk '{print $3}')",     //Удаляем изображение
		// "docker system prune -a -f",
		"GOOS=linux go build -o ./gibdd ./main.go",
		"make base",
		"rm -f ./gibdd", //Удаляем исходинк
	}
	// log.Println("Собираем архив")
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
