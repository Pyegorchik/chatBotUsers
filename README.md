# Telegram Chat Parser Bot

Телеграм-бот для извлечения списка участников из экспорта истории чата.

## Что делает

- Принимает файлы экспорта Telegram (JSON/HTML)
- Извлекает уникальных участников
- Формирует список (<50 участников) или Excel файл (≥50 участников)
- Не хранит данные пользователей

## Быстрый старт

### 1. Получите токен бота
```bash
# Напишите @BotFather в Telegram
# /newbot → следуйте инструкциям
```

### 2. Запуск

**С Go:**
```bash
go mod download
export TELEGRAM_BOT_TOKEN="your-token"
go run .
```

**С Docker:**
```bash
echo "TELEGRAM_BOT_TOKEN=your-token" > .env
docker-compose up -d
```

### 3. Использование

1. Экспортируйте историю чата (Telegram Desktop → Меню → Экспорт → JSON)
2. Отправьте файл(ы) боту
3. Отправьте `/process`
4. Получите результат

## Команды бота

- `/start` - Начать работу
- `/help` - Справка
- `/clear` - Очистить файлы
- `/process` - Обработать файлы

## Структура проекта

```
.
├── main.go              # Точка входа
├── handler.go           # Обработчик команд
├── parser.go            # Парсер JSON/HTML
├── excel.go             # Генератор Excel
├── parser_test.go       # Тесты
├── go.mod               # Зависимости
├── Dockerfile           # Docker
├── docker-compose.yml   # Docker Compose
└── examples/            # Примеры файлов
```

## Скриншоты работы
<img width="579" height="502" alt="image" src="https://github.com/user-attachments/assets/b8b7159b-b314-441c-82c1-c5392ef73137" />
<img width="516" height="188" alt="image" src="https://github.com/user-attachments/assets/101561d7-4e20-4e62-bce1-0584f5f3919e" />
<img width="518" height="560" alt="image" src="https://github.com/user-attachments/assets/bb9cf7aa-5741-4b99-a045-2c5fde2cce45" />
<img width="702" height="266" alt="image" src="https://github.com/user-attachments/assets/fed0c34b-dd04-477b-86f8-14b950242e28" />


