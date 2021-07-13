module gibdd

go 1.15

replace gibdd/telegram => ./telegram

replace gibdd/utils => ./utils

require (
	gibdd/telegram v0.0.0-00010101000000-000000000000
	gibdd/utils v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.7
)
