package main

import (
	"GETSEPOLIA/internal/browser"
	"GETSEPOLIA/internal/captcha"
	"log"
	"time"
)

func main() {
	ethAddress := "0x9ca0f5b88b3e06e1e9e34d39af353f55ddcd45f5"
	proxy := "res.proxy-sale.com:10000"

	ctx, cancel := browser.NewBrowserContext(proxy)
	defer cancel()

	err := browser.NavigateToURL(ctx, "https://sepolia-faucet.pk910.de/")
	if err != nil {
		log.Fatalf("Помилка переходу на сайт: %v", err)
	}

	err = browser.EnterETHAddress(ctx, ethAddress)
	if err != nil {
		log.Fatalf("Не вдалося ввести адресу: %v", err)
	}

	captchaType, err := browser.DetectCaptchaType(ctx)
	if err != nil {
		log.Fatalf("Не вдалося визначити тип капчі: %v", err)
	}

	if captchaType != browser.CaptchaTypeHCaptcha {
		log.Fatalf("Потрібна hCaptcha, а не %s", captchaType)
	}

	captchaToken, _, err := captcha.SolveHCaptcha(proxy)
	if err != nil {
		log.Fatalf("Не вдалося вирішити капчу: %v", err)
	}

	err = browser.InjectCaptchaToken(ctx, captchaToken)
	if err != nil {
		log.Fatalf("Помилка вставки капча-токену: %v", err)
	}

	err = browser.ClickStartMining(ctx)
	if err != nil {
		log.Fatalf("Не вдалося натиснути кнопку Start Mining: %v", err)
	}

	log.Println("Кнопку натиснуто. Очікуємо на відкриття вікна майнінгу...")
	time.Sleep(30 * time.Second)
}


//"res.proxy-sale.com:10000"