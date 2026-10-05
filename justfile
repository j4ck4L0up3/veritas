default-db := "./data/dev.db"

add-sql name target=default-db:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} create {{name}} sql

up target=default-db:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} up

down target=default-db:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} down

build:
  go build -o ./veritas ./cmd/...

gen:
  sqlc generate
