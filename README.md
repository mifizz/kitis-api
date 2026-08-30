# kitis-api

## Установка и запуск
Для начала ставим postgresql, если ещё не, потом создаём юзера kitis и базу kitisapi:
```bash
sudo -u postgres psql
```
```sql
CREATE USER kitis WITH PASSWORD 'тут_пароль_юзера';
CREATE DATABASE kitisapi OWNER kitis;
```

Клонируем этот репо куда нужно:
```bash
git clone https://github.com/mifizz/kitis-api
cd kitis-api
```

Копируем `config/example.yaml` в `config/default.yaml` и при необходимости меняем значения в конфиге.
```bash
cp config/example.yaml config/default.yaml
```

Запускаем через go:
```bash
go run cmd/api/main.go
```

Либо билдим и потом запускаем бинарник:
```bash
go build cmd/api/main.go
./main
```

## Документация
Документация доступна (или будет доступа) на https://shkitis.ru/api
