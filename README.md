# Subscriptions API

Проект предоставляет REST API для управления подписками пользователей.
Спецификация API доступна в файле swagger.yaml (можно открыть в редакторе: https://editor.swagger.io/)

# Возможности

+ Создание, обновление и удаление подписок
+ Получение информации о подписке по ID
+ Список подписок пользователя
+ Подсчёт общей стоимости активных подписок за выбранный диапазон месяцев

# Пример .env файла (расположить в корне проекта)

```
POSTGRES_USER=user
POSTGRES_PASSWORD=passwordfordb
POSTGRES_DB=substracker_db
POSTGRES_HOST=db
POSTGRES_PORT=5432
APP_PORT=8081
```

# Запуск

```bash
docker-compose up --build
```

Сервис будет доступен по адресу http://localhost:8081