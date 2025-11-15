# prservice-task
Test task for internship(Pull Request service)


├── cmd
│   └── app
│       └── main.go               # точка входа здесь или в интернал?
│
├── domain                        # доменные сущности
│   ├── team.go                   # Team
│   ├── user.go                   # User
│   ├── pull_request.go           # PullRequest
│
├── dto                           # нужны ли они тут вообще?
│   ├── team_dto.go               # Team
│   ├── user_dto.go               # User
│   └── pull_request_dto.go       # PullRequest
│
├── internal
│   ├── app
│   │   └── app.go                # я так понимаю тут сборка всех сервисов и зависимостей идет
│   │
│   ├── config
│   │   └── config.go             # в studyproject было и в целом по стилистике го есть, нужен ли в тестовом мне?
│   │
│   ├── repository                # работа с domain-моделями
│   │   ├── team_repository.go    # TeamRepository
│   │   ├── user_repository.go    # UserRepository
│   │   └── pull_request_repository.go  # PullRequestRepository
│   │
│   ├── service                   # бизнес-логика
│   │   ├── team_service.go       # CreateTeam, GetTeam
│   │   ├── user_service.go       # SetIsActive, GetUser
│   │   └── pull_request_service.go  # CreatePR, MergePR, ReassignReviewer, GetUserReviews
│   │
│   └── web                       # вот тут немного не понял, как все должно быть организовано, взял чисто, как гпт объяснил, но все равно не уверен, что это верно
│       └── http
│           ├── router.go         # инициализация mux, навешивание хендлеров и middleware
│           ├── handlers_team.go  # /team/add, /team/get
│           ├── handlers_user.go  # /users/setIsActive, /users/getReview
│           └── handlers_pr.go    # /pullRequest/create, /pullRequest/merge, /pullRequest/reassign
│
├── pkg
│   └── postgres
│       ├── postgres.go           # инициализация подключения к Postgres (dsn, retries, логгер)
│       └── gorm.go               # если пользуешься GORM: функция NewGormDB(...) *gorm.DB
│
├── migrations
│   ├── 000001_teams.up.sql
│   ├── 000001_teams.down.sql
│   ├── 000002_users.up.sql
│   ├── 000002_users.down.sql
│   ├── 000003_pull_requests.up.sql
│   └── 000003_pull_requests.down.sql
│
├── openapi
│   └── openapi.yaml              # спецификация из задания (так понял надо оставить без изменений или все-таки надо как-то менять?)
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md

