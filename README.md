Телеграм бот по проверке штрафов gibdd

Оcобенности:

Своя реализация телеграм бота на основе публичного api

БД mysql для хранения данных

Использование YandexDisk для резервирования БД и раскат приложения с имеющимея данными на любом сервере


# gibdd

.bash_profile

export telegaGibddToken=XXXX

export myIDtelega=XXXX //chat_id для дебажных сообщений

##yandex-disk setup https://yandex.ru/support/disk-desktop-linux/start.html

export iidYandex=XXXX

export passwdYandex=XXXX

export remoteHost="111.111.111.111" ##Хост с докером для установки приложения

export remoteUser="username" ##Имя пользователя на удаленном хосте

source .profile

env

myIDtelega=XXXX

telegaGibddToken=XXXX

id_rsa в корне с приватным ключем доступа в облако
