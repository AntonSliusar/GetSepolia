package captcha

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	SolveCaptchaAPIKey   = "4c60bf018dbbca7b3460090b521c418d"
	HCaptchaSiteKey       = "89693841-2505-4039-8c39-479c9188991f"
	HCaptchaPageURL       = "https://sepolia-faucet.pk910.de/"
	SolveCaptchaBaseURL   = "https://api.solvecaptcha.com"
)

type solveCaptchaInResponse struct {
	Status    int    `json:"status"`
	Request   string `json:"request"`
	ErrorText string `json:"error_text"`
}

type solveCaptchaResResponse struct {
	Status     int    `json:"status"`
	Request    string `json:"request"`
	UserAgent  string `json:"useragent"`
	ErrorText  string `json:"error_text"`
}

func SolveHCaptcha(proxy string) (string, string, error) {
	log.Println("Розв'язуємо hCaptcha через SolveCaptcha API...")

	data := url.Values{}
	data.Set("key", SolveCaptchaAPIKey)
	data.Set("method", "hcaptcha")
	data.Set("pageurl", HCaptchaPageURL)
	data.Set("sitekey", HCaptchaSiteKey)
	data.Set("json", "1")

	client := &http.Client{}
	if proxy != "" {
		proxyURL, err := url.Parse("http://" + proxy)
		if err != nil {
			return "", "", fmt.Errorf("неправильна адреса проксі: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
		}
	}

	resp, err := client.PostForm(SolveCaptchaBaseURL+"/in.php", data)
	if err != nil {
		return "", "", fmt.Errorf("не вдалося відправити завдання hCaptcha: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var inResp solveCaptchaInResponse
	if err := json.Unmarshal(body, &inResp); err != nil {
		return "", "", fmt.Errorf("помилка JSON: %w; відповідь: %s", err, string(body))
	}
	if inResp.Status != 1 {
		return "", "", fmt.Errorf("помилка створення завдання: %s", inResp.Request)
	}

	captchaID := inResp.Request
	log.Printf("Завдання створено. ID: %s", captchaID)

	for i := 0; i < 24; i++ {
		time.Sleep(5 * time.Second)
		url := fmt.Sprintf("%s/res.php?key=%s&action=get&id=%s&json=1", SolveCaptchaBaseURL, SolveCaptchaAPIKey, captchaID)
		res, err := client.Get(url)
		if err != nil {
			return "", "", fmt.Errorf("помилка при запиті результату: %w", err)
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		var resResp solveCaptchaResResponse
		if err := json.Unmarshal(body, &resResp); err != nil {
			return "", "", fmt.Errorf("JSON помилка: %w. Тіло: %s", err, string(body))
		}
		if resResp.Status == 1 {
			log.Println("Капча успішно розв'язана.")
			return resResp.Request, resResp.UserAgent, nil
		}
		if resResp.Request != "CAPCHA_NOT_READY" {
			return "", "", fmt.Errorf("помилка при отриманні токену: %s", resResp.Request)
		}
		log.Println("Очікуємо на рішення...")
	}
	return "", "", fmt.Errorf("час очікування рішення капчі вичерпано")
}