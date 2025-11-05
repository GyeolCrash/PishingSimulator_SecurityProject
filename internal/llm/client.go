package llm

import (
	"PishingSimulator_SecurityProject/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const llmBaseURL = "http://localhost:8000" // LLM 서버의 기본 URL
var httpClient = &http.Client{Timeout: 10 * time.Second}

type InitRequest struct {
	SessionID   string             `json:"session_id"`
	Scenario    string             `json:"scenario"`
	UserInfo    models.UserProfile `json:"user_info"`
	Temperature float64            `json:"temperature"`
}

type InitResponse struct {
	Utterance string `json:"utterance"`
}

type ChatRequest struct {
	SessionID string `json:"session_id"`
	Utterance string `json:"user_text"`
}

type ChatResponse struct {
	Utterance string `json:"utterance"`
	NextStep  string `json:"next_step"`
}

type ControlRequest struct {
	ClearSession bool `json:"clear_session"`
}

func InitSession(sessionID, scenarioKey string, userInfo models.UserProfile) (string, error) {
	reqBody, err := json.Marshal(InitRequest{
		SessionID:   sessionID,
		Scenario:    scenarioKey,
		UserInfo:    userInfo,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(llmBaseURL+"/init", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("LLM Server init failed with status: " + resp.Status)
	}

	var initResp InitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return "", err
	}
	return initResp.Utterance, nil
}

func Chat(sessionID, text string) (*ChatResponse, error) {
	reqBody, err := json.Marshal(ChatRequest{
		SessionID: sessionID,
		Utterance: text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Post(llmBaseURL+"/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("LLM Server chat failed with status: " + resp.Status)
	}
	return &chatResp, nil
}

func ClearSession(sessionID string) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"clear_session": true,
	})
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(llmBaseURL+"/control", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("LLM session %s cleared.", sessionID)
	return nil
}
