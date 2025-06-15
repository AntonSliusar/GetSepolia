package main

import (
	"GETSEPOLIA/internal/browser"
	"GETSEPOLIA/internal/captcha"
	"log"
	"time"
)

const (
	TargetURL = "https://sepolia-faucet.pk910.de/"
)

func main() {

	// // wallet := "0x0079EA086BA71b8DBbb341033458a3a38eEcB42B"
	// proxyAddr := ""
	// // antiCaptchaAPIKey := "98c5510fb5661c0511a3371de51c6e35"
	// faucetPageURL := "https://sepolia-faucet.pk910.de/#/"

	// antiCaptchaAPIKey := "98c5510fb5661c0511a3371de51c6e35"
	// ethAddress := "0x9ca0f5b88b3e06e1e9e34d39af353f55ddcd45f5" 
	// //placeholder := "Please enter ETH address or ENS name"           // потім додати в конфіг
	// hardcodedRecaptchaSiteKey := "6Leg_psiAAAAAHlE_PSnJuYLQDXbrnBw6G2l_vvu"

// Створюємо контекст браузера
	TestProxyAddress :=  ""//"http://198.23.239.134:6540"
	TestETHAddress := "0x0079ea086ba71b8dbbb341033458a3a38eecb42b" // Приклад адреси


	ctx, cancel := browser.NewBrowserContext(TestProxyAddress)
	defer cancel() // Забезпечуємо закриття контексту при завершенні main

    time.Sleep(5 * time.Second) // Спробуйте 5 секунд, якщо не допоможе, збільшіть до 10-15
	err := browser.NavigateToURL(ctx, TargetURL)
	if err != nil {
		log.Fatalf("Failed to navigate to %s: %v", TargetURL, err)
	}
	
	err = browser.EnterETHAddress(ctx, TestETHAddress)
	if err != nil {
		log.Fatalf("Failed to enter ETH address: %v", err)
	}

	recaptchaToken, err := captcha.SolveRecaptcha() // Тепер викликаємо SolveRecaptcha
	if err != nil {
		log.Fatalf("Failed to solve reCAPTCHA: %v", err)
	}

	err = browser.SetRecaptchaResponse(ctx, recaptchaToken) // Тепер викликаємо SetRecaptchaResponse
	if err != nil {
		log.Fatalf("Failed to set reCAPTCHA response: %v", err)
	}

	log.Println("Successfully opened the browser and navigated to the target URL. Waiting for 30 seconds to observe...")
	time.Sleep(300 * time.Second) // Зачекаємо 30 секунд, щоб візуально переконатись
	log.Println("Closing browser.")	
	
}