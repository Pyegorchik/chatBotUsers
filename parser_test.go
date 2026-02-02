package main

import (
	"fmt"
	"os"
	"testing"
)

func TestParser_ParseJSON(t *testing.T) {
	parser := NewParser()

	jsonData := []byte(`{
		"messages": [
			{
				"type": "message",
				"from": "John Doe",
				"from_id": "user1001",
				"text": "Hello"
			},
			{
				"type": "message",
				"from": "Jane Smith",
				"from_id": "user1002",
				"text": "Hi"
			},
			{
				"type": "message",
				"from": "John Doe",
				"from_id": "user1001",
				"text": "How are you?"
			},
			{
				"type": "message",
				"from": "",
				"from_id": "",
				"text": "Deleted message"
			}
		]
	}`)

	err := parser.parseJSON(jsonData)
	if err != nil {
		t.Fatalf("Неудалось спарсить result.json: %v", err)
	}

	participants := parser.GetParticipants()

	// Проверяем количество участников (должно быть 2, так как John Doe дублируется и Deleted Account исключен)
	if len(participants) != 2 {
		t.Errorf("Ожидалось 2 участника, получили %d", len(participants))
	}

	// Проверяем наличие участников по ID
	foundJohn := false
	foundJane := false

	for _, p := range participants {
		fmt.Printf("Participants %v\n", p)
		if p.ID == "user1001" {
			foundJohn = true
			if p.FirstName != "John" {
				t.Errorf("Ожидалось FirstName 'John', получили '%s'", p.FirstName)
			}
			if p.LastName != "Doe" {
				t.Errorf("Ожидалось LastName 'Doe', получили '%s'", p.LastName)
			}
		}
		if p.ID == "user1002" {
			foundJane = true
			if p.FirstName != "Jane" {
				t.Errorf("Ожидалось FirstName 'Jane', получили '%s'", p.FirstName)
			}
			if p.LastName != "Smith" {
				t.Errorf("Ожидалось LastName 'Smith', получили '%s'", p.LastName)
			}
		}
	}

	if !foundJohn {
		t.Error("John Doe not found in participants")
	}
	if !foundJane {
		t.Error("Jane Smith not found in participants")
	}
}

func TestParser_IsJSON(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{"Valid JSON object", `{"key": "value"}`, true},
		{"Valid JSON array", `[1, 2, 3]`, true},
		{"JSON with whitespace", `  {"key": "value"}  `, true},
		{"HTML data", `<!DOCTYPE html><html></html>`, false},
		{"Plain text", `This is plain text`, false},
		{"Empty string", ``, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isJSON([]byte(tt.data))
			if result != tt.expected {
				t.Errorf("isJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParser_DuplicateFiltering(t *testing.T) {
	parser := NewParser()

	// Добавляем одного и того же участника трижды
	jsonData := []byte(`{
		"messages": [
			{
				"id": 1,
				"type": "message",
				"from": "Test User",
				"from_id": "1001",
				"text": "Message 1"
			},
			{
				"id": 2,
				"type": "message",
				"from": "Test User",
				"from_id": "1001",
				"text": "Message 2"
			},
			{
				"id": 3,
				"type": "message",
				"from": "Test User",
				"from_id": "1001",
				"text": "Message 3"
			}
		]
	}`)

	err := parser.parseJSON(jsonData)
	if err != nil {
		t.Fatalf("Неудалось спарсить result.json: %v", err)
	}

	participants := parser.GetParticipants()

	// Должен быть только один участник
	if len(participants) != 1 {
		t.Errorf("Ожидался 1, получили %d", len(participants))
	}

	if len(participants) > 0 {
		if participants[0].FirstName != "Test" {
			t.Errorf("Ожидалось FirstName 'Test', получили '%s'", participants[0].FirstName)
		}
		if participants[0].LastName != "User" {
			t.Errorf("Ожидалось LastName 'User', получили '%s'", participants[0].LastName)
		}
	}
}

func TestParser_DeletedAccountFiltering(t *testing.T) {
	parser := NewParser()

	jsonData := []byte(`{
		"messages": [
			{
				"id": 1,
				"type": "message",
				"from": "Deleted Account",
				"from_id": "0",
				"text": "Deleted message"
			},
			{
				"id": 2,
				"type": "message",
				"from": "Valid User",
				"from_id": "1001",
				"text": "Valid message"
			}
		]
	}`)

	err := parser.parseJSON(jsonData)
	if err != nil {
		t.Fatalf("Неудалось спарсить result.json: %v", err)
	}

	participants := parser.GetParticipants()

	// Должен быть только один участник (deleted account отфильтрован)
	if len(participants) != 1 {
		t.Errorf("Ожидался один участник, а получили %d", len(participants))
	}

	if len(participants) > 0 {
		if participants[0].FirstName != "Valid" {
			t.Errorf("Ожидалось 'Valid', получили '%s'", participants[0].FirstName)
		}
		if participants[0].LastName != "User" {
			t.Errorf("Ожидалось LastName 'User', получили '%s'", participants[0].LastName)
		}
	}
}

// Тест для загрузки локального JSON файла
func TestParser_ParseLocalJSON(t *testing.T) {
	// Проверяем существование файла
	if _, err := os.Stat("result.json"); os.IsNotExist(err) {
		t.Skip("Файл result.json не найден, пропускаем тест")
		return
	}

	data, err := os.ReadFile("result.json")
	if err != nil {
		t.Fatalf("Неудалось прочитать result.json: %v", err)
	}

	parser := NewParser()
	err = parser.parseJSON(data)
	if err != nil {
		t.Fatalf("Неудалось спарсить result.json: %v", err)
	}

	participants := parser.GetParticipants()

	fmt.Println("\n=== Участники из result.json ===")
	fmt.Printf("Всего участников: %d\n\n", len(participants))

	for i, p := range participants {
		fullName := p.FirstName
		if p.LastName != "" {
			fullName += " " + p.LastName
		}
		fmt.Printf("%d. %s (ID: %s)\n", i+1, fullName, p.ID)
	}
	fmt.Println()

	// Создаем Excel файл
	if len(participants) > 0 {
		excelPath, err := GenerateExcel(participants)
		if err != nil {
			t.Errorf("Не удалось создать Excel: %v", err)
		} else {
			fmt.Printf("✅ Excel файл создан: %s\n", excelPath)
		}
	}
}

// Тест для загрузки локального HTML файла
func TestParser_ParseLocalHTML(t *testing.T) {
	// Проверяем существование файла
	if _, err := os.Stat("messages.html"); os.IsNotExist(err) {
		t.Skip("Файл messages.html не найден, пропускаем тест")
		return
	}

	data, err := os.ReadFile("messages.html")
	if err != nil {
		t.Fatalf("Не удалось прочитать messages.html: %v", err)
	}

	parser := NewParser()
	err = parser.parseHTML(data)
	if err != nil {
		t.Fatalf("Неудалось спарсить messages.html: %v", err)
	}

	participants := parser.GetParticipants()

	fmt.Println("\n=== Участники из messages.html ===")
	fmt.Printf("Всего участников: %d\n\n", len(participants))

	for i, p := range participants {
		fullName := p.FirstName
		if p.LastName != "" {
			fullName += " " + p.LastName
		}
		fmt.Printf("%d. %s (ID: %s)\n", i+1, fullName, p.ID)
	}
	fmt.Println()

	// Создаем Excel файл
	if len(participants) > 0 {
		excelPath, err := GenerateExcel(participants)
		if err != nil {
			t.Errorf("Неудалось сгенерировать Excel: %v", err)
		} else {
			fmt.Printf("✅ Excel файл создан: %s\n", excelPath)
		}
	}
}
