package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(participants []Participant) (string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Участники"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}

	f.SetActiveSheet(index)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return "", err
	}

	headers := []string{"№", "Дата экспорт1а", "Имя", "Фамилия", "Описание"}
	for i, header := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", colName)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	f.SetColWidth(sheetName, "A", "A", 5)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 40)

	exportDate := time.Now().Format("2006-01-02 15:04:05")

	for i, participant := range participants {
		row := i + 2 // Начинаем со второй строки (первая - заголовки)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), exportDate)

		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), participant.FirstName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), participant.LastName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), participant.Bio)
	}

	// Удаляем стандартный лист Sheet1 если он есть
	if f.GetSheetName(0) == "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	tmpDir := os.TempDir()
	fileName := fmt.Sprintf("participants_%d.xlsx", time.Now().Unix())
	filePath := filepath.Join(tmpDir, fileName)

	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}
