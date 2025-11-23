# prservice-task
Test task for internship(PR Reviewer Assignment Service)
Сервис назначения ревьюеров для Pull Request’ов.

В данном репозитории реализован сервис для автоматического назначения ревьюверов на Pull Request’ы внутри команд разработки. Реализован на Go 1.23, использует PostgreSQL, Docker Compose, миграции и oapi-codegen.

## Запуск приложения
Для запуска и поднятия приложения в терминале выполните команду: 
```bash
docker-compose up
```
После приложение запустится и сервис будет доступен на: http://localhost:8080. База данных поднимется автоматически, миграции применятся при запуске контейнера приложения.
Для остановки: 
```bash
docker-compose down
```
Для полной очистки:
```bash
docker-compose down -v
```
Переменные окружения:
- `PORT` - порт HTTP (по умолчанию: :8080)
- `DB_HOST` - хост БД (по умолчанию: db)
- `DB_PORT` - порт БД (по умолчанию: :5432)
- `DB_USER` - пользователь (по умолчанию: postgres)
- `DB_PASSWORD` - пароль (по умолчанию: PostgresPass)
- `DB_NAME` - имя БД (по умолчанию: PostgresPass)
- `MIGRATE_ENABLE` - авто-миграции (по умолчанию: true)
- `MIGRATE_FOLDER` - путь к миграциям (по умолчанию: ./migrations)

## Makefile команды

### Кодогенерация

```bash
make generate-openapi        # oapi-кодогенерация
```

### Миграции 

```bash
make migrate        # Ручное применение миграций
```

```bash
make migrate-down-all        # Ручной откат всех миграций
```

### Локальный запуск приложения

```bash
make run
```

### Тестирование

```bash
make load-test        # Запуск нагрузочного тестирование (для запуска требуется k6)
```

### Линтер

```bash
make lint         # Запуск golangci-lint
```

## API

Полная спецификация — openapi.yml

Основные эндпоинты:
### Teams

- `POST /team/add` - cоздать команду 
- `GET /team/get?team_name=X` - получить команду
- `POST /team/deactivate` - массовая деактивация + безопасное переназначение PR

### Users

- `POST /users/setIsActive` - изменить активность пользователя
- `GET /users/getReview?user_id=X` - PR, где пользователь ревьювер

### Pull Requests

- `POST /pullRequest/create` - cоздать PR с автоматическим назначением ревьюверов
- `POST /pullRequest/merge` - идемпотентный merge
- `POST /pullRequest/reassign` - переназначить ревьювера

### Stats

- `GET /stats` - вся суммарная статистика

## Архитектура приложения

Архитектура проекта была создана согласно принципам Clean Architecture

├── cmd
│   └── app
│       └── main.go                       # точка входа в приложениеточка входа: грузим конфиг, создаём App
│
├── config                                # порт, DSN, прочие настройки
│   └── config.go 
│
├── internal
│   ├── app
│   │   └── app.go                        # сборка всех зависимостей (db, repos, services, router)
│   │
│   ├── errs                              # работа с ошибками
│   │   └── errors.go                     # Errors
│   │
│   ├── models                            # работа с моделями
│   │   ├── team.go                       # Team
│   │   ├── user.go                       # User
│   │   ├── stats.go                      # Stats
│   │   └── pullrequest.go                # PullRequest
│   │
│   ├── repository                        # работа с репозиториями
│   │   ├── repository.go                 # контракты
│   │   └── postgres
│   │       ├── team.go                       # TeamRepository
│   │       ├── user.go                       # UserRepository
│   │       ├── stats.go                      # StatsRepository
│   │       └── pullrequest.go                # PullRequestRepository
│   │
│   ├── service                        # бизнес-логика
│   │   ├── team.go                    # CreateTeam, GetTeamByName, DeactivateUsersAndReassignPRs
│   │   ├── user.go                    # SetFlagIsActive, GetActiveUsersByTeam
│   │   ├── stats.go                   # GetStats
│   │   └── pull.go                    # CreatePullRequest, MergePullRequest, GetPullRequestsByReviewer, ReassignToPullRequest
│   │
│   └── web                            # HTTP-слой
│       ├── codes.go                   # конвертация доменных ошибок в JSON-ответы
│       ├── middleware.go              # логгирование
│       ├── router.go                  # регистрация всех маршрутов
│       ├── routers-pull-requests.go   # хендлеры для пулл реквестов
│       ├── router-teams.go            # хендлеры для команд
│       ├── router-users.go            # хендлеры для пользователей
│       ├── router-stats.go            # хендлеры для статистики
│       └── omodels
│           └── api.gen.go             # сгенерированные OpenAPI-модели
│
├── pkg
│   └── postgres
│       └── postgres.go                # инициализация подключения к Postgres
│
├── migrations                         # миграции
│   ├── 000001_teams.up.sql
│   ├── 000001_teams.down.sql
│   ├── 000002_users.up.sql
│   ├── 000002_users.down.sql
│   ├── 000003_pull_requests.up.sql
│   └── 000003_pull_requests.down.sql
│
├── api
│   └── openapi.yml                  # спецификация OpenAPI
├── .golangci.yml                    # конфигурация линтера
├── k6.js                            # load-tests
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.sum
├── go.mod
└── README.md

## Тестирование

В рамках данного приложения удалось выполнить нагрузочное тестирование:

- `make load-test` - запустить нагрузочное тестирование (необходим k6 и запущенный сервер в другом терминале)

Результаты нагрузочного тестирования:

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

     execution: local
        script: ./k6.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 20 looping VUs for 30s (gracefulStop: 30s)

  █ TOTAL RESULTS

    checks_total.......: 8721    285.833467/s
    checks_succeeded...: 100.00% 8721 out of 8721
    checks_failed......: 0.00%   0 out of 8721

    ✓ team created or already exists
    ✓ create PR: status 201
    ✓ getReview: status 200
    ✓ merge PR: status 200

    HTTP
    http_req_duration..............: avg=28.29ms  min=1.81ms   med=20.31ms  max=332.96ms p(90)=58.22ms  p(95)=75.21ms
      { expected_response:true }...: avg=28.29ms  min=1.81ms   med=20.31ms  max=332.96ms p(90)=58.22ms  p(95)=75.21ms
    http_req_failed................: 0.00%  0 out of 8721
    http_reqs......................: 8721   285.833467/s

    EXECUTION
    iteration_duration.............: avg=172.41ms min=109.78ms med=161.61ms max=466.21ms p(90)=228.26ms p(95)=257.52ms
    iterations.....................: 3486   114.254726/s
    vus............................: 20     min=20        max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 783 MB 26 MB/s
    data_sent......................: 1.6 MB 51 kB/s

running (0m30.5s), 00/20 VUs, 3486 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  30s

### Анализ результатов:

Профиль нагрузки: 20 виртуальных пользователей, длительность теста - 30 секунд.
Требования из ТЗ: RPS — 5, SLI времени ответа — 300 мс, SLI успешности — 99.9%.

По результам теста мы видим, cервис стабильно выдерживает нагрузку порядка ~285 RPS при 20 одновременных пользователях.
SLI по времени ответа (p95 < 300 ms) соблюдён: p95 ≈ 75 ms.
SLI по успешности (99.9% успешных запросов) также выполняется.

