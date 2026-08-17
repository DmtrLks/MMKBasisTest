# Task Manager API Для ММК- Базис.

REST API для управления командами и задачами. Проект реализован как тестовое задание на Go с MySQL, Redis, JWT-аутентификацией, историей изменений, комментариями и SQL-отчетом.

## Быстрый запуск через Docker Compose

Требуется Docker с Compose v2 (команда `docker compose`, не `docker-compose`).

```bash
docker compose --env-file .env.docker up -d --build
```

Флаг `--env-file .env.docker` для конфигурации по умолчанию не обязателен: в `docker-compose.yml` у всех подстановок заданы значения по умолчанию, повторяющие этот файл, поэтому `docker compose up -d --build` на свежем клоне тоже работает. Флаг нужен, если вы правите в `.env.docker` порт или доступы к MySQL: сам контейнер API читает файл всегда, а вот подстановки в `docker-compose.yml` без флага возьмут значения по умолчанию, и настройки разъедутся.

Если порты на хосте заняты, поменяйте их в `.env.docker`: `APP_PORT` задает порт API сразу и в контейнере, и на хосте, а `DB_HOST_PORT` и `REDIS_HOST_PORT` меняют только публикацию MySQL и Redis наружу, не влияя на подключение приложения к ним.

**Перед любым использованием за пределами локальной разработки замените `JWT_SECRET`.** В `.env.docker` и `.env.example` лежит значение для ознакомительного запуска: оно есть в репозитории, поэтому с ним можно подписать валидный токен на любого пользователя.

Миграции выполняются приложением автоматически.

Проверка:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Локальный запуск API

Поднять только инфраструктуру:

```bash
docker compose --env-file .env.docker up -d mysql redis
```

Создать локальный env-файл.

Bash:

```bash
cp .env.example .env
```

В `.env.example` уже используются `localhost:3306` и `localhost:6379`. После настройки `JWT_SECRET` запустить API:

```bash
go run ./cmd/api
```

## Конфигурация

Приложение читает переменные окружения и при локальном запуске пытается загрузить `.env`.

| Переменная | Назначение | Пример |
|---|---|---|
| `APP_PORT` | HTTP-порт API (в контейнере и на хосте) | `8080` |
| `RATE_LIMIT_RPS` | Допустимое число запросов в секунду на IP | `10` |
| `RATE_LIMIT_BURST` | Максимальный кратковременный всплеск запросов на IP | `20` |
| `DB_HOST`, `DB_PORT` | Адрес MySQL, к которому подключается API | `localhost`, `3306` |
| `DB_HOST_PORT` | Порт публикации MySQL на хост (только Compose) | `3306` |
| `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Доступ к MySQL | `task_manager`, `app`, `app_password` |
| `DB_MAX_OPEN_CONNS` | Максимум открытых соединений | `25` |
| `DB_MAX_IDLE_CONNS` | Максимум idle-соединений | `10` |
| `DB_CONN_MAX_LIFETIME` | Максимальное время жизни соединения | `30m` |
| `DB_CONN_MAX_IDLE_TIME` | Максимальное idle-время | `5m` |
| `REDIS_HOST`, `REDIS_PORT` | Адрес Redis, к которому подключается API | `localhost`, `6379` |
| `REDIS_HOST_PORT` | Порт публикации Redis на хост (только Compose) | `6379` |
| `REDIS_PASSWORD`, `REDIS_DB` | Доступ и номер Redis DB | пусто, `0` |
| `TASK_LIST_CACHE_TTL` | TTL кеша списка задач | `5m` |
| `JWT_SECRET` | Ключ подписи JWT | должен быть заменен |
| `JWT_ISSUER` | Issuer JWT | `task-manager` |
| `JWT_TTL` | Время жизни JWT | `24h` |
| `HTTP_*_TIMEOUT` | Таймауты HTTP-сервера | см. `.env.example` |

Duration-переменные используют формат Go: `500ms`, `10s`, `5m`, `24h`.

## API

Полный контракт находится в [docs/swagger.yaml](docs/swagger.yaml).

| Метод | Endpoint | Назначение |
|---|---|---|
| `GET` | `/health` | Healthcheck |
| `POST` | `/api/v1/register` | Регистрация |
| `POST` | `/api/v1/login` | Получение JWT |
| `POST` | `/api/v1/teams` | Создание команды |
| `GET` | `/api/v1/teams` | Команды пользователя |
| `POST` | `/api/v1/teams/{id}/invite` | Приглашение участника |
| `PATCH` | `/api/v1/teams/{id}/members/{user_id}` | Изменение роли участника |
| `POST` | `/api/v1/tasks` | Создание задачи |
| `GET` | `/api/v1/tasks` | Список задач с фильтрами |
| `PUT` | `/api/v1/tasks/{id}` | Обновление задачи |
| `POST` | `/api/v1/tasks/{id}/comments` | Добавление комментария |
| `GET` | `/api/v1/tasks/{id}/comments` | Комментарии задачи |
| `GET` | `/api/v1/tasks/{id}/history` | История задачи |
| `GET` | `/api/v1/teams/{team_id}/stats` | SQL-отчет команды |

Все endpoint'ы, кроме health, регистрации и login, требуют заголовок:

```http
Authorization: Bearer <access_token>
```

### Примеры использвания:

Регистрация:

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com","password":"password123","name":"Owner"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com","password":"password123"}'
```

Создание команды:

```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Backend"}'
```

Изменение роли участника (только owner):

```bash
curl -X PATCH http://localhost:8080/api/v1/teams/1/members/5 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
```

Создание задачи:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"team_id":1,"title":"Implement API","status":"todo"}'
```

Получение задач:

```bash
curl "http://localhost:8080/api/v1/tasks?team_id=1&status=todo&limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```


Если задача уже была изменена и версия устарела, API вернет `409 Conflict`.

Получение статистики команды:

```bash
curl http://localhost:8080/api/v1/teams/1/stats \
  -H "Authorization: Bearer $TOKEN"
```

Пример ответа:

```json
{
  "team_id": 1,
  "tasks_by_status": {
    "todo": 4,
    "in_progress": 2,
    "done": 7
  },
  "top_assignees": [
    {
      "user_id": 5,
      "name": "Ivan Petrov",
      "closed_tasks": 3
    }
  ],
  "average_close_time_seconds": 86400,
  "comments_count": 12
}
```

## Права доступа

- `owner` и `admin` могут приглашать участников, изменять любые задачи команды и смотреть отчет;
- менять роли участников может только `owner`; роль `owner` выдать нельзя, и роль самого создателя команды неизменна;
- создатель задачи может изменять название, описание, статус и исполнителя своей задачи;
- исполнитель может изменять только статус назначенной ему задачи;
- обычный участник не может изменять чужую задачу;
- задачи, история и комментарии недоступны пользователям другой команды;
- новый исполнитель должен состоять в той же команде;
- роль `owner` нельзя выдать ни через приглашение, ни через смену роли.

## Rate limiting

Для каждого IP-адреса создается отдельный token bucket. `RATE_LIMIT_RPS` задает скорость восстановления токенов, а `RATE_LIMIT_BURST` — максимальный кратковременный всплеск. При превышении лимита API отвечает `429 Too Many Requests` с заголовком `Retry-After`. Эндпоинт `/health` не ограничивается.

## SQL-отчет

`GET /api/v1/teams/{team_id}/stats` доступен только owner/admin и возвращает:

- количество задач по статусам;
- top-3 исполнителей по закрытым задачам за последние 30 дней;
- среднее время закрытия в секундах;
- количество комментариев по задачам команды.

Отчет выполняется одним SQL-запросом и фильтруется по `team_id`.

## Тесты

Интеграционный тест SQL-отчета использует MySQL и выполняет подготовку данных внутри транзакции с rollback.
