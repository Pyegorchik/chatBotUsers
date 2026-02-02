package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot       *tgbotapi.BotAPI
	userFiles map[int64][]string // chatID -> список file_id
	mu        sync.RWMutex
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		bot:       bot,
		userFiles: make(map[int64][]string),
	}
}

func (h *Handler) HandleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			h.handleStart(chatID)
		case "help":
			h.handleHelp(chatID)
		case "clear":
			h.handleClear(chatID)
		case "process":
			h.handleProcess(chatID)
		default:
			h.sendMessage(chatID, "Неизвестная команда. Используйте /help для просмотра доступных команд.")
		}
		return
	}

	if message.Document != nil {
		h.handleDocument(chatID, message.Document)
		return
	}

	// Обработка обычного текста
	h.sendMessage(chatID, "Пожалуйста, отправьте файл истории чата (JSON или HTML) или используйте команду /help для помощи.")
}

func (h *Handler) handleStart(chatID int64) {
	welcomeMsg := `👋 Добро пожаловать в бот для извлечения участников чата!

Этот бот поможет вам получить список участников из экспорта истории Telegram чата.

📝 Как использовать:
1. Экспортируйте историю чата в Telegram (Настройки → Экспорт данных)
2. Отправьте файл(ы) экспорта боту (не более 10 файлов)
3. Используйте команду /process для обработки

Бот поддерживает форматы: JSON и HTML
Ваши данные обрабатываются на лету и не сохраняются на сервере.

Используйте /help для получения подробной информации.`

	h.sendMessage(chatID, welcomeMsg)
}

func (h *Handler) handleHelp(chatID int64) {
	helpMsg := `📚 Доступные команды:

/start - Начать работу с ботом
/help - Показать это сообщение
/clear - Очистить загруженные файлы
/process - Обработать загруженные файлы

📋 Как использовать бота:

1. Экспортируйте историю чата:
   - Откройте чат в Telegram Desktop
   - Меню → Экспорт истории чата
   - Выберите формат JSON или HTML
   
2. Отправьте файлы боту:
   - Максимум 10 файлов за раз
   - Поддерживаются: .json, .html
   
3. Запустите обработку командой /process

📊 Результат:
- Если участников < 50: список в чате
- Если участников ≥ 50: Excel файл

🔒 Конфиденциальность:
Файлы обрабатываются мгновенно и не сохраняются на сервере.`

	h.sendMessage(chatID, helpMsg)
}

func (h *Handler) handleClear(chatID int64) {
	h.mu.Lock()
	delete(h.userFiles, chatID)
	h.mu.Unlock()

	h.sendMessage(chatID, "✅ Все загруженные файлы очищены. Можете загрузить новые файлы.")
}

func (h *Handler) handleDocument(chatID int64, document *tgbotapi.Document) {
	fileName := strings.ToLower(document.FileName)
	if !strings.HasSuffix(fileName, ".json") && !strings.HasSuffix(fileName, ".html") {
		h.sendMessage(chatID, "❌ Неподдерживаемый формат файла. Пожалуйста, отправьте JSON или HTML файл.")
		return
	}

	h.mu.Lock()
	fileList := h.userFiles[chatID]

	if len(fileList) >= MaxFiles {
		h.mu.Unlock()
		h.sendMessage(chatID, fmt.Sprintf("❌ Достигнут лимит файлов (%d). Используйте /clear для очистки или /process для обработки.", MaxFiles))
		return
	}

	fileList = append(fileList, document.FileID)
	h.userFiles[chatID] = fileList
	h.mu.Unlock()

	h.sendMessage(chatID, fmt.Sprintf("✅ Файл '%s' добавлен (%d/%d). Отправьте еще файлы или используйте /process для обработки.", document.FileName, len(fileList), MaxFiles))
}

func (h *Handler) handleProcess(chatID int64) {
	h.mu.RLock()
	fileIDs, exists := h.userFiles[chatID]
	h.mu.RUnlock()

	if !exists || len(fileIDs) == 0 {
		h.sendMessage(chatID, "❌ Нет загруженных файлов. Пожалуйста, сначала отправьте файлы истории чата.")
		return
	}

	h.sendMessage(chatID, fmt.Sprintf("⏳ Обрабатываю %d файл(ов)... Пожалуйста, подождите.", len(fileIDs)))

	participants, err := h.processFiles(fileIDs)
	if err != nil {
		log.Printf("Ошибка подготовки файла для %d: %v", chatID, err)
		h.sendMessage(chatID, fmt.Sprintf("❌ Ошибка при обработке файлов: %v", err))
		return
	}

	h.mu.Lock()
	delete(h.userFiles, chatID)
	h.mu.Unlock()

	if len(participants) < 50 {
		h.sendParticipantsList(chatID, participants)
	} else {
		h.sendParticipantsExcel(chatID, participants)
	}
}

func (h *Handler) processFiles(fileIDs []string) ([]Participant, error) {
	parser := NewParser()

	for _, fileID := range fileIDs {
		fileURL, err := h.bot.GetFileDirectURL(fileID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения URL: %w", err)
		}

		if err := parser.ParseFileFromURL(fileURL); err != nil {
			return nil, fmt.Errorf("ошибка парсинга файла: %w", err)
		}
	}

	return parser.GetParticipants(), nil
}

func (h *Handler) sendParticipantsList(chatID int64, participants []Participant) {
	if len(participants) == 0 {
		h.sendMessage(chatID, "ℹ️ Участники не найдены.")
		return
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("👥 Найдено участников: %d\n\n", len(participants)))

	for i, p := range participants {
		username := p.FirstName
		if p.LastName != "" {
			username += " " + p.LastName
		}

		message.WriteString(fmt.Sprintf("%d. %s\n", i+1, username))
	}

	h.sendMessage(chatID, message.String())
}

func (h *Handler) sendParticipantsExcel(chatID int64, participants []Participant) {
	excelPath, err := GenerateExcel(participants)
	if err != nil {
		log.Printf("Ошибка генерации Excel: %v", err)
		h.sendMessage(chatID, fmt.Sprintf("❌ Ошибка при создании Excel файла: %v", err))
		return
	}

	msg := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(excelPath))
	msg.Caption = fmt.Sprintf("📊 Список участников (%d человек)", len(participants))

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки Excel файла: %v", err)
		h.sendMessage(chatID, "❌ Ошибка при отправке файла.")
	}
}

func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
