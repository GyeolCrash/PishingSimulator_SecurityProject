package llm

import (
	"PishingSimulator_SecurityProject/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const llmBaseURL = "http://localhost:8001" // LLM 서버의 기본 URL
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
	UserText  string `json:"user_text"`
}

type ChatResponse struct {
	Utterance string `json:"utterance"`
	NextStep  string `json:"next_step"`
}

type ControlRequest struct {
	ClearSession bool `json:"clear_session"`
}

func InitSession(sessionID, scenarioKey string, userInfo models.UserProfile, ctx context.Context) (string, error) {
	reqBody, err := json.Marshal(InitRequest{
		SessionID:   sessionID,
		Scenario:    scenarioKey,
		UserInfo:    userInfo,
		Temperature: 0.7,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", llmBaseURL+"/session/init", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Context-Type", "application/json")

	resp, err := httpClient.Do(req)
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

func Chat(sessionID, text string, ctx context.Context) (*ChatResponse, error) {
	reqBody, err := json.Marshal(ChatRequest{
		SessionID: sessionID,
		UserText:  text,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", llmBaseURL+"/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Context-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("LLM Server chat failed with status: " + resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	return &chatResp, nil
}

func ClearSession(sessionID string) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"session_id":    sessionID,
		"clear_session": true,
	})
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(llmBaseURL+"/session/control", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("LLM session %s cleared.", sessionID)
	return nil
}
