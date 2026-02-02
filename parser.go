package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Participant struct {
	// Username  string
	FirstName string
	LastName  string
	Bio       string
	ID        string
}

type Parser struct {
	participants map[string]Participant // ID -> Participant
}

func NewParser() *Parser {
	return &Parser{
		participants: make(map[string]Participant),
	}
}

func (p *Parser) ParseFileFromURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ошибка скачивания файла: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if p.isJSON(data) {
		return p.parseJSON(data)
	}
	return p.parseHTML(data)
}

func (p *Parser) isJSON(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func (p *Parser) parseJSON(data []byte) error {
	var export TelegramExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	for _, msg := range export.Messages {
		// Пропускаем служебные сообщения и удаленные аккаунты
		if msg.From == "" || msg.From == "Deleted Account" || msg.FromID == "" {
			continue
		}

		participant := Participant{
			ID: msg.FromID,
		}

		parts := strings.SplitN(msg.From, " ", 2)
		participant.FirstName = parts[0]
		if len(parts) > 1 {
			participant.LastName = parts[1]
		}

		p.participants[msg.FromID] = participant
	}

	return nil
}

func (p *Parser) parseHTML(data []byte) error {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("ошибка парсинга HTML: %w", err)
	}

	p.extractFromHTML(doc)
	return nil
}

func (p *Parser) extractFromHTML(n *html.Node) {
	// Ищем div с классом "message" или "from_name"
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "message") {
				p.extractMessageFromHTML(n)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.extractFromHTML(c)
	}
}

func (p *Parser) extractMessageFromHTML(n *html.Node) {
	var fromName string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "from_name") {
					fromName = p.getTextContent(node)
					break
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)

	if fromName != "" && fromName != "Deleted Account" {
		// Создаем простого участника (ID генерируем из имени для дедупликации)
		id := fromName

		participant := Participant{
			ID: id,
		}

		parts := strings.SplitN(fromName, " ", 2)
		participant.FirstName = parts[0]
		if len(parts) > 1 {
			participant.LastName = parts[1]
		}

		p.participants[id] = participant
	}
}

func (p *Parser) getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			return strings.TrimSpace(c.Data)
		}
		// text += p.getTextContent(c)
	}
	return strings.TrimSpace(text)
}

func (p *Parser) GetParticipants() []Participant {
	result := make([]Participant, 0, len(p.participants))
	for _, participant := range p.participants {
		result = append(result, participant)
	}
	return result
}

func hashString(s string) uint32 {
	h := uint32(0)
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

type TelegramExport struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	// ID     string `json:"id"`
	Type   string `json:"type"`
	Date   string `json:"date"`
	From   string `json:"from"`
	FromID string `json:"from_id"`
	// FromName string `json:"from_name,omitempty"`
	// Text string `json:"text"`
}
