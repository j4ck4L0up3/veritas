default-target := "./data/dev.db"

add-sql name target=default-target:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} create {{name}} sql

up target=default-target:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} up

down target=default-target:
  ~/go/bin/goose sqlite3 -dir ./db/migrations {{target}} down
