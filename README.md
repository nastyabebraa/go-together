# Ride Together Bot

Telegram-бот для поиска попутчиков в каршеринге.

## Запуск

Требуются Go 1.25 и MySQL.

```shell
cp .env.example .env
set -a
source .env
set +a
go run .
```

Обязательные переменные окружения перечислены в `.env.example`. Идентификаторы стикеров необязательны.
