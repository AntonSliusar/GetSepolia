package main

import (
	"GETSEPOLIA/internal/browser"
	"GETSEPOLIA/internal/captcha"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	const maxRetries = 3
	ethAddress := "0x9ca0f5b88b3e06e1e9e34d39af353f55ddcd45f5"
		proxy := "95.134.179.170:59100"

		ctx, cancel := browser.NewBrowserContext(proxy)
		defer cancel()

	for attempt := 1; attempt <= maxRetries; attempt++ {

		if attempt == 1 {
			if err := browser.NavigateToURL(ctx, "https://sepolia-faucet.pk910.de/"); err != nil {
				log.Fatalf("Перехід на сторінку: %w", err)
			}
		} else {
			if err := browser.RefreshPage(ctx); err != nil {
				log.Fatalf("не вдалося оновити сторінку: %w", err)
			}
		}

		err := browser.EnterETHAddress(ctx, ethAddress)
		if err != nil {
			log.Fatalf("Не вдалося ввести адресу: %v", err)
		}

		//Solve CAPTCHA
		captchaType, err := browser.DetectCaptchaType(ctx)
		if err != nil {
			log.Fatalf("Не вдалося визначити тип капчі: %v", err)
		}
		if captchaType == browser.CaptchaTypeHCaptcha {
			captchaToken, _, err := captcha.SolveHCaptcha(proxy)
			if err != nil {
				log.Fatalf("Не вдалося вирішити капчу: %v", err)
			}

			err = browser.InjectCaptchaToken(ctx, captchaToken)
			if err != nil {
				log.Fatalf("Помилка вставки капча-токену: %v", err)
			}
		}
		if captchaType == browser.CaptchaTypeRecaptcha {
			captchaToken, err := captcha.SolveReCAPTCHA()
			if err != nil {
				log.Fatalf("Не вдалося вирішити капчу: %v", err)
				continue
			}
			err = chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('textarea[name="g-recaptcha-response"]').value = '%s'`, captchaToken), nil),
			chromedp.Evaluate(fmt.Sprintf(`document.getElementById('g-recaptcha-response').innerHTML = '%s'`, captchaToken), nil),
			chromedp.Sleep(2*time.Second),
			)
			if err != nil {
				log.Fatalf("Не вдалося вставити токен капчі: %v", err)
			}
		}

		

		err = browser.ClickStartMining(ctx)
		time.Sleep(3 * time.Second)
		if err != nil {
			log.Fatalf("Не вдалося натиснути кнопку Start Mining: %v", err)
		}

		invalideCaptcha, _ := browser.CheckForInvalidCaptchaError(ctx)
		if invalideCaptcha {
				log.Println("[INVALID_CAPTCHA] — повторюємо спробу...")
				continue
		}
		log.Println("Кнопку натиснуто. Очікуємо на відкриття вікна майнінгу...")
		time.Sleep(100 * time.Second)
		
	}	

	log.Println("Кнопку натиснуто. Очікуємо на відкриття вікна майнінгу...")
	time.Sleep(30 * time.Second)
}


