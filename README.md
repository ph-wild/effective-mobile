# REST-сервис для агрегации информации о подписках пользователей.
Реализованы CRUDL-операций над записями о подписках. Выставлена HTTP-ручка для подсчета суммарной стоимости всех подписок за выбранный период с фильтрацией по id пользователя и названию подписки  
  
## Стек:
Язык программирования: Go, миграции - Goose, логи - slog  
Протокол запросов: HTTP (через chi)  
База данных: PostgreSQL (sqlx)  
Контейнеризация: Docker Compose  
  
## Запуск сервиса
```
docker-compose up -d       # 1. Запускаем PostgreSQL  
make build                 # 2. Собираем бинарник приложения  
make migrate-up            # 3. Применяем миграции к БД  
make run                   # 4. Запускаем приложение  
```  

Альтернатива (все в одной команде): ```make up```

  
Swagger будет доступен по адресу:  
http://localhost:8080/swagger/index.html  


## Архитектура
Сервис разделён на три основных слоя:  
handler — HTTP-ручки (работают с chi)  
service — бизнес-логика  
storage — взаимодействие с базой через sqlx  

Все зависимости передаются в main.go (Dependency Injection)

### Структура проекта:
```
.  
├── cmd/  
│   └── main.go                     # Точка входа: инициализация зависимостей и запуск  
├── config/  
│   └── config.go                   # Загрузка конфигурации из YAML  
├── internal/  
│   ├── db/connect.go               # Подключение к PostgreSQL через sqlx  
│   ├── models/subscription.go      # Структуры Subscription и SubscriptionInput  
│   ├── storage/subscription.go     # SQL-запросы: CRUD + Summary  
│   ├── service/subscription.go     # Бизнес-логика с интерфейсом  
│   └── handler/  
│       ├── subscription.go         # HTTP-хендлеры для API и регистрация всех роутов  
│       └── middleware.go           # Регистрация всех роутов + middleware  
├── migrations/                     # Goose-миграции  
├── docs/                           # Сгенерированный swag (swagger.json, swagger.yaml, docs.go)  
├── .gitignore                      # Игнорирование файлов  
├── docker-compose.yml              # PostgreSQL-сервис  
├── config.yaml                     # Конфигурационный файл (порт, DSN и пр.)  
├── Makefile                        # Сборка, запуск, миграции, swag  
└── README.md                       # Документация проекта  
```

На текущем этапе сервис не использует транзакции по следующим причинам:
- Каждая операция (INSERT, UPDATE, DELETE) выполняется над одной таблицей.  
- Нет связанных сущностей и необходимости атомарных операций.  
- SQL-база (PostgreSQL) по умолчанию гарантирует согласованность одиночных запросов.  

### Интерфейсы  
Слой service реализует интерфейс SubscriptionServiceInterface, это позволяет:
- Упростить mock для тестов  
- Легко заменить реализацию бизнес-логики  
- Реализация скрыта за структурой subscriptionService, которая создаётся через конструктор NewSubscriptionService(...)  

### Тесты
В задании не указано покрытие тестами, поэтому оно отсутствует.  