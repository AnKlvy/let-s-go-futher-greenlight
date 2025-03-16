# Зависимости
go get github.com/julienschmidt/httprouter
go get github.com/lib/pq
go get golang.org/x/time/rate

# Запуск Docker
docker-compose up -d

# Управление PostgreSQL через Docker
Подключение от имени пользователя postgres

docker exec -it postgres_greenlight psql -U postgres -d greenlight

Подключение от имени пользователя greenlight

docker exec -it postgres_greenlight psql -U greenlight -d greenlight

# Создание миграций
migrate create -seq -ext .sql -dir ./migrations create_movies_table \
migrate create -seq -ext .sql -dir ./migrations add_movies_check_constraints
migrate create -seq -ext .sql -dir ./migrations add_movies_indexes

# Применение миграций
# С использованием переменной окружения
migrate -path ./migrations -database "$env:GREENLIGHT_DB_DSN" up



# С явным указанием строки подключения
migrate -path ./migrations -database "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable" up

Отключаем SSL для локальной работы

# Откат миграций до определенной версии
migrate -path ./migrations -database $EXAMPLE_DSN goto 1

# Тест для optimistic locking (PowerShell)
 1..8 | ForEach-Object -Parallel {
    Invoke-RestMethod -Uri "http://localhost:4000/v1/movies/6" -Method Patch -Body '{"runtime": "97 mins"}' -ContentType "application/json"
 } -ThrottleLimit 8

 # Тест для rate limiting
  1..6 | ForEach-Object { Invoke-WebRequest -Uri "http://localhost:4000/v1/healthcheck" }
