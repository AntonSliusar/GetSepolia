package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MiningRequest struct {
	Address      string `json:"addr"`
	CaptchaToken string `json:"captchaToken"`
}

func StartMiningSession(address, captchaToken, userAgent string) error {
	log.Println("Відправляємо запит на старт майнінгу...")

	requestBody := MiningRequest{
		Address:      address,
		CaptchaToken: captchaToken,
	}

	encodedBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("не вдалося закодувати JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "https://sepolia-faucet.pk910.de/api/startSession", bytes.NewBuffer(encodedBody))
	if err != nil {
		return fmt.Errorf("не вдалося створити запит: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://sepolia-faucet.pk910.de/")
	req.Header.Set("Origin", "https://sepolia-faucet.pk910.de")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("помилка запиту: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер повернув код: %d", resp.StatusCode)
	}

	log.Println("Майнінг сесію успішно запущено.")
	return nil
}
