package session

// import (
// 	"GETSEPOLIA/internal/api"
// 	"GETSEPOLIA/internal/browser"
// 	"GETSEPOLIA/internal/captcha"
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"
// )

// func AttemptStartMining(ctx context.Context, ethAddress, proxy string) error {
// 	log.Println("Переходимо на сайт...")
// 	if err := browser.NavigateToURL(ctx, "https://sepolia-faucet.pk910.de/"); err != nil {
// 		return err
// 	}

// 	if err := browser.EnterETHAddress(ctx, ethAddress); err != nil {
// 		return err
// 	}

// 	captchaType, err := browser.DetectCaptchaType(ctx)
// 	if err != nil {
// 		return fmt.Errorf("не вдалося визначити тип капчі: %w", err)
// 	}

// 	switch captchaType {
// 	case browser.CaptchaTypeHCaptcha:
// 		log.Println("Розв'язуємо hCaptcha...")
// 		token, userAgent, err := captcha.SolveHCaptcha(proxy)
// 		if err != nil {
// 			return fmt.Errorf("не вдалося розв'язати hCaptcha: %w", err)
// 		}

// 		if err := api.StartMiningSession(ethAddress, token, userAgent); err != nil {
// 			return fmt.Errorf("помилка при старті сесії з hCaptcha: %w", err)
// 		}

// 	case browser.CaptchaTypeRecaptcha:
// 		log.Println("Розв'язуємо reCAPTCHA...")
// 		token, userAgent, err := captcha.SolveRecaptcha()
// 		if err != nil {
// 			return fmt.Errorf("не вдалося розв'язати reCAPTCHA: %w", err)
// 		}

// 		if err := api.StartMiningSession(ethAddress, token, userAgent); err != nil {
// 			return fmt.Errorf("помилка при старті сесії з reCAPTCHA: %w", err)
// 		}

// 	default:
// 		return fmt.Errorf("не виявлено підтримуваної капчі")
// 	}

// 	log.Println("Майнінг сесію успішно запущено! Очікуємо перед закриттям браузера...")
// 	time.Sleep(30 * time.Second)
// 	return nil
// }
