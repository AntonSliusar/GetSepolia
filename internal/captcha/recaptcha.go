package captcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// Ключі API для сервісу Anti-captcha
	AntiCaptchaAPIKey = "98c5510fb5661c0511a3371de51c6e35" // Ваш Anti-captcha API ключ

	// Ключ сайту для reCAPTCHA
	RecaptchaSiteKey = "6Leg_psiAAAAAHlE_PSnJuYLQDXbrnBw6G2l_vvu" // Sitekey для reCAPTCHA V2 з логів

	// URL-адреси для взаємодії з сервісом Anti-captcha та цільовим сайтом
	RecaptchaPageURL   = "https://sepolia-faucet.pk910.de/"
	AntiCaptchaBaseURL = "https://api.anti-captcha.com"
)

type createTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      struct {
		Type        string `json:"type"`
		WebsiteURL  string `json:"websiteURL"`
		WebsiteKey  string `json:"websiteKey"`
		IsInvisible bool   `json:"isInvisible,omitempty"`
	} `json:"task"`
}

type createTaskResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	TaskID           int    `json:"taskId"`
}

type getTaskResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    int    `json:"taskId"`
}

type getTaskResultResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	Status           string `json:"status"` // "processing" or "ready"
	Solution         struct {
		GRecaptchaResponse string `json:"gRecaptchaResponse"`
	} `json:"solution"`
}

func SolveReCAPTCHA() (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Створення запиту для створення завдання
	createReq := createTaskRequest{
		ClientKey: AntiCaptchaAPIKey,
	}
	createReq.Task.Type = "RecaptchaV2TaskProxyless"
	createReq.Task.WebsiteURL = RecaptchaPageURL
	createReq.Task.WebsiteKey = RecaptchaSiteKey
	createReq.Task.IsInvisible = true // оскільки сайт використовує невидиму капчу

	// Відправка запиту на створення завдання
	jsonBody, _ := json.Marshal(createReq)
	resp, err := client.Post("https://api.anti-captcha.com/createTask", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("createTask error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var createResp createTaskResponse
	json.Unmarshal(body, &createResp)

	if createResp.ErrorID != 0 {
		return "", fmt.Errorf("anti-captcha error: %s - %s", createResp.ErrorCode, createResp.ErrorDescription)
	}

	// Опитування до моменту, поки завдання не буде вирішене
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Second)

		getReq := getTaskResultRequest{
			ClientKey: AntiCaptchaAPIKey,
			TaskID:    createResp.TaskID,
		}
		jsonBody, _ := json.Marshal(getReq)

		resp, err := client.Post("https://api.anti-captcha.com/getTaskResult", "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			return "", fmt.Errorf("getTaskResult error: %v", err)
		}
		defer resp.Body.Close()

		body, _ = io.ReadAll(resp.Body)
		var getResp getTaskResultResponse
		json.Unmarshal(body, &getResp)

		if getResp.ErrorID != 0 {
			return "", fmt.Errorf("getTaskResult error: %s - %s", getResp.ErrorCode, getResp.ErrorDescription)
		}

		if getResp.Status == "ready" {
			return getResp.Solution.GRecaptchaResponse, nil
		}

		log.Printf("Очікуємо розв'язання капчі... %d сек", (i+1)*5)
	}

	return "", fmt.Errorf("капча не вирішена протягом очікування")
}