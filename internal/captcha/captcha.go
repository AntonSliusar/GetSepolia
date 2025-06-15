package captcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	AntiCaptchaAPIKey  = "98c5510fb5661c0511a3371de51c6e35" // Ваш Anti-captcha API ключ
	RecaptchaSiteKey    = "6Leg_psiAAAAAHlE_PSnJuYLQDXbrnBw6G2l_vvu" // Sitekey для reCAPTCHA V2 з ваших логів
	RecaptchaPageURL    = "https://sepolia-faucet.pk910.de/"  
	AntiCaptchaBaseURL = "https://api.anti-captcha.com"
)

// createTaskRequest структура для запиту створення завдання
type createTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      struct {
		Type       string `json:"type"` // "NoCaptchaTaskProxyless" для reCAPTCHA V2
		WebsiteURL string `json:"websiteURL"`
		WebsiteKey string `json:"websiteKey"`
	} `json:"task"`
}

// createTaskResponse структура для відповіді створення завдання
type createTaskResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	TaskID           int    `json:"taskId"`
}

// getTaskResultRequest структура для запиту результату завдання
type getTaskResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    int    `json:"taskId"`
}

type getTaskResultResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	Status           string `json:"status"` // "processing" або "ready"
	Solution         struct {
		GRecaptchaResponse string `json:"gRecaptchaResponse"` // Поле для reCAPTCHA токена
	} `json:"solution"`
}

func SolveRecaptcha() (string, error) {
	log.Println("Solving reCaptcha using Anti-captcha...")

	// 1. Створення завдання
	createReqBody := createTaskRequest{
		ClientKey: AntiCaptchaAPIKey,
		Task: struct {
			Type       string `json:"type"`
			WebsiteURL string `json:"websiteURL"`
			WebsiteKey string `json:"websiteKey"`
		}{
			Type:       "NoCaptchaTaskProxyless", // Тип для reCAPTCHA V2
			WebsiteURL: RecaptchaPageURL,
			WebsiteKey: RecaptchaSiteKey,
		},
	}

	createReqJSON, err := json.Marshal(createReqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create task request: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/createTask", AntiCaptchaBaseURL), "application/json", bytes.NewBuffer(createReqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to send create task request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read create task response body: %w", err)
	}

	var createResp createTaskResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal create task response: %w. Response body: %s", err, string(body))
	}

	if createResp.ErrorID != 0 {
		return "", fmt.Errorf("anti-captcha error creating task (%s): %s", createResp.ErrorCode, createResp.ErrorDescription)
	}
	if createResp.TaskID == 0 {
		return "", fmt.Errorf("anti-captcha returned empty task ID: %s", string(body))
	}

	log.Printf("Anti-captcha task created, ID: %d. Waiting for result...", createResp.TaskID)

	// 2. Опитування результату завдання
	getReqBody := getTaskResultRequest{
		ClientKey: AntiCaptchaAPIKey,
		TaskID:    createResp.TaskID,
	}
	getReqJSON, err := json.Marshal(getReqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal get task result request: %w", err)
	}

	// Опитуємо результат кожні 5 секунд, максимум 120 секунд
	for i := 0; i < 24; i++ { // 24 * 5 секунд = 120 секунд (2 хвилини)
		time.Sleep(5 * time.Second)

		resp, err := http.Post(fmt.Sprintf("%s/getTaskResult", AntiCaptchaBaseURL), "application/json", bytes.NewBuffer(getReqJSON))
		if err != nil {
			return "", fmt.Errorf("failed to send get task result request: %w", err)
		}
		defer resp.Body.Close() // Дефер буде спрацьовувати на кожній ітерації, це ОК.

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read get task result response body: %w", err)
		}

		var getResp getTaskResultResponse
		if err := json.Unmarshal(body, &getResp); err != nil {
			return "", fmt.Errorf("failed to unmarshal get task result response: %w. Response body: %s", err, string(body))
		}

		if getResp.ErrorID != 0 {
			return "", fmt.Errorf("anti-captcha error getting task result (%s): %s", getResp.ErrorCode, getResp.ErrorDescription)
		}

		if getResp.Status == "ready" {
			log.Println("hCaptcha solved successfully by Anti-captcha!")
			return getResp.Solution.GRecaptchaResponse, nil
		} else if getResp.Status == "processing" {
			log.Printf("hCaptcha still processing... attempt %d", i+1)
		} else {
			return "", fmt.Errorf("anti-captcha returned unknown status: %s. Response body: %s", getResp.Status, string(body))
		}
	}

	return "", fmt.Errorf("anti-captcha failed to solve hCaptcha within timeout")
}