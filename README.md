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

# Применение миграций
# С использованием переменной окружения
migrate -path ./migrations -database "$env:GREENLIGHT_DB_DSN" up



# С явным указанием строки подключения
migrate -path ./migrations -database "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable" up

Отключаем SSL для локальной работы

# Откат миграций до определенной версии
migrate -path ./migrations -database $EXAMPLE_DSN goto 1
