module gibdd

go 1.15

replace gibdd/telegram => ./telegram

require (
	gibdd/telegram v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.7
)
