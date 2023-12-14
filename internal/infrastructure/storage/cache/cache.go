package cache

import (
	"BaumanS2SBot/internal/model"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

var mutex sync.Mutex

type FileDataList struct {
	Requests []model.FileData `json:"requests"`
}

func LoadRequests(filename string) (FileDataList, error) {

	var data FileDataList

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("No file finded: %v", err)
			return data, nil
		}
		return data, err
	}

	if len(fileBytes) == 0 {
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

func DeleteExpiredRequestsFromCache(filename string, loc *time.Location) (error, map[int64][]int, map[int64][]int) {

	data, err := LoadRequests(filename)
	if err != nil {
		return err, nil, nil
	}
	messageIDToDelete := make(map[int64][]int)
	messageIDToEdit := make(map[int64][]int)
	var validRequests []model.FileData
	for _, req := range data.Requests {
		if time.Now().In(loc).Before(req.ExpiryDate) {
			validRequests = append(validRequests, req)
		} else if time.Now().In(loc).After(req.ExpiryDate) && req.OrigMessageID != 0 { // чтобы было московское время добавляем 3 часа
			messageIDToDelete[req.ChatID] = append(messageIDToDelete[req.ChatID], req.ForwardMessageID)
		} else if time.Now().In(loc).After(req.ExpiryDate) && req.OrigMessageID == 0 {
			messageIDToEdit[req.ChatID] = append(messageIDToDelete[req.ChatID], req.ForwardMessageID)
			log.Print(req.ChatID, req.ForwardMessageID)
		}
	}

	data.Requests = validRequests
	return SaveRequests(filename, data), messageIDToDelete, messageIDToEdit
}

func DeleteRequest(filename string, messageID int) (map[int64][]int, map[int64]map[int]bool, error) {
	data, err := LoadRequests(filename)
	if err != nil {
		return nil, nil, err
	}

	deleteMap := make(map[int64][]int)
	editMap := make(map[int64]map[int]bool)

	var validRequests []model.FileData
	for _, req := range data.Requests {
		if req.OrigMessageID == 0 {
			if editMap[req.ChatID] == nil {
				editMap[req.ChatID] = make(map[int]bool)
			}
			editMap[req.ChatID][req.ForwardMessageID] = req.IsMedia
		}

		if messageID != req.OrigMessageID {
			validRequests = append(validRequests, req)
		} else {
			deleteMap[req.ChatID] = append(deleteMap[req.ChatID], req.ForwardMessageID)
		}
	}

	data.Requests = validRequests

	err = SaveRequests(filename, data)
	if err != nil {
		return nil, nil, err
	}

	return deleteMap, editMap, nil
}

func AddRequest(filename string, chatID int64, messageID int, expiryDate time.Time, forwardMessageID int, isMedia bool) error {
	data, err := LoadRequests(filename)
	if err != nil {
		return err
	}

	newRequest := model.FileData{
		ChatID:           chatID,
		OrigMessageID:    messageID,
		ExpiryDate:       expiryDate,
		ForwardMessageID: forwardMessageID,
		IsMedia:          isMedia,
	}
	data.Requests = append(data.Requests, newRequest)
	return SaveRequests(filename, data)
}
