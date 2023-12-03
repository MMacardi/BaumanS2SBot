package cache

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

var mutex sync.Mutex

type FileData struct {
	ChatID           int64     `json:"chat_id"`
	MessageID        int       `json:"message_id"`
	ExpiryDate       time.Time `json:"expiry_date"`
	ForwardMessageID int       `json:"forward_message_id"`
}

type FileDataList struct {
	Requests []FileData `json:"requests"`
}

func LoadRequests(filename string) (FileDataList, error) {

	var data FileDataList

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, возвращаем пустой список
			log.Fatalf("No file finded: %v", err)
			return data, nil
		}
		return data, err
	}

	if len(fileBytes) == 0 {
		// Если файл пустой, возвращаем пустой список
		log.Print("File is empty")
		return data, nil
	}
	err = json.Unmarshal(fileBytes, &data)
	return data, err
}

func SaveRequests(filename string, data FileDataList) error {
	fileBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, fileBytes, 0644)
}

// DeleteExpiredRequests удаляет устаревшие запросы
func DeleteExpiredRequests(filename string, loc *time.Location) (error, map[int64][]int) {

	data, err := LoadRequests(filename)
	if err != nil {
		return err, nil
	}
	messageIDToDelete := make(map[int64][]int)
	var validRequests []FileData
	for _, req := range data.Requests {
		if time.Now().In(loc).Before(req.ExpiryDate) {
			validRequests = append(validRequests, req)
		} else if time.Now().In(loc).After(req.ExpiryDate) { // чтобы было московское время добавляем 3 часа
			//messageIDToDelete[req.ChatID] = append(messageIDToDelete[req.ChatID], req.MessageID)
			messageIDToDelete[req.ChatID] = append(messageIDToDelete[req.ChatID], req.ForwardMessageID)
		}
	}

	data.Requests = validRequests
	return SaveRequests(filename, data), messageIDToDelete
}

func DeleteRequest(filename string, messageID int) (error, map[int64][]int) {

	data, err := LoadRequests(filename)
	if err != nil {
		return err, nil
	}
	deleteMap := make(map[int64][]int)
	var validRequests []FileData
	for _, req := range data.Requests {
		if messageID != req.MessageID {
			validRequests = append(validRequests, req)
		} else {
			deleteMap[req.ChatID] = append(deleteMap[req.ChatID], req.ForwardMessageID)
		}
	}

	data.Requests = validRequests
	return SaveRequests(filename, data), deleteMap
}

// AddRequest добавляет новый запрос в файл
func AddRequest(filename string, chatID int64, messageID int, expiryDate time.Time, forwardMessageID int) error {
	data, err := LoadRequests(filename)
	if err != nil {
		return err
	}

	newRequest := FileData{
		ChatID:           chatID,
		MessageID:        messageID,
		ExpiryDate:       expiryDate,
		ForwardMessageID: forwardMessageID,
	}
	data.Requests = append(data.Requests, newRequest)
	return SaveRequests(filename, data)
}
