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

// createTaskRequest структура для запиту створення завдання в Anti-captcha
type createTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      struct {
		Type       string `json:"type"` // "NoCaptchaTaskProxyless" для reCAPTCHA V2
		WebsiteURL string `json:"websiteURL"`
		WebsiteKey string `json:"websiteKey"`
	} `json:"task"`
}

// createTaskResponse структура для відповіді створення завдання від Anti-captcha
type createTaskResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	TaskID           int    `json:"taskId"`
}

// getTaskResultRequest структура для запиту результату завдання від Anti-captcha
type getTaskResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    int    `json:"taskId"`
}

// getTaskResultResponse структура для відповіді результату завдання від Anti-captcha
type getTaskResultResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	Status           string `json:"status"` // "processing" або "ready"
	Solution         struct {
		GRecaptchaResponse string `json:"gRecaptchaResponse"` // Поле для reCAPTCHA токена
	} `json:"solution"`
}

// SolveRecaptcha розв'язує reCAPTCHA за допомогою сервісу Anti-captcha.
func SolveRecaptcha() (string, error) {
	log.Println("Розв'язуємо reCAPTCHA за допомогою Anti-captcha...")

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
		return "", fmt.Errorf("не вдалося маршалізувати запит створення завдання: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/createTask", AntiCaptchaBaseURL), "application/json", bytes.NewBuffer(createReqJSON))
	if err != nil {
		return "", fmt.Errorf("не вдалося відправити запит створення завдання: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("не вдалося прочитати тіло відповіді створення завдання: %w", err)
	}

	var createResp createTaskResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return "", fmt.Errorf("не вдалося демаршалізувати відповідь створення завдання: %w. Тіло відповіді: %s", err, string(body))
	}

	if createResp.ErrorID != 0 {
		return "", fmt.Errorf("помилка Anti-captcha під час створення завдання (%s): %s", createResp.ErrorCode, createResp.ErrorDescription)
	}
	if createResp.TaskID == 0 {
		return "", fmt.Errorf("Anti-captcha повернув пустий ID завдання: %s", string(body))
	}

	log.Printf("Завдання Anti-captcha створено, ID: %d. Очікуємо на результат...", createResp.TaskID)

	// 2. Опитування результату завдання
	getReqBody := getTaskResultRequest{
		ClientKey: AntiCaptchaAPIKey,
		TaskID:    createResp.TaskID,
	}
	getReqJSON, err := json.Marshal(getReqBody)
	if err != nil {
		return "", fmt.Errorf("не вдалося маршалізувати запит на отримання результату завдання: %w", err)
	}

	// Опитуємо результат кожні 5 секунд, максимум 120 секунд
	for i := 0; i < 24; i++ { // 24 * 5 секунд = 120 секунд (2 хвилини)
		time.Sleep(5 * time.Second)

		resp, err := http.Post(fmt.Sprintf("%s/getTaskResult", AntiCaptchaBaseURL), "application/json", bytes.NewBuffer(getReqJSON))
		if err != nil {
			return "", fmt.Errorf("не вдалося відправити запит на отримання результату завдання: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("не вдалося прочитати тіло відповіді отримання результату завдання: %w", err)
		}

		var getResp getTaskResultResponse
		if err := json.Unmarshal(body, &getResp); err != nil {
			return "", fmt.Errorf("не вдалося демаршалізувати відповідь отримання результату завдання: %w. Тіло відповіді: %s", err, string(body))
		}

		if getResp.ErrorID != 0 {
			return "", fmt.Errorf("помилка Anti-captcha під час отримання результату завдання (%s): %s", getResp.ErrorCode, getResp.ErrorDescription)
		}

		if getResp.Status == "ready" {
			log.Println("reCAPTCHA успішно розв'язана за допомогою Anti-captcha!")
			return getResp.Solution.GRecaptchaResponse, nil
		} else if getResp.Status == "processing" {
			log.Printf("reCAPTCHA все ще обробляється... спроба %d", i+1)
		} else {
			return "", fmt.Errorf("Anti-captcha повернув невідомий статус: %s. Тіло відповіді: %s", getResp.Status, string(body))
		}
	}

	return "", fmt.Errorf("Anti-captcha не вдалося розв'язати reCAPTCHA протягом встановленого часу")
}