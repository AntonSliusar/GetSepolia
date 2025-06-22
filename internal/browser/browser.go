package browser

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// Константа для типів капчі
const (
	CaptchaTypeNone     = "none"
	CaptchaTypeRecaptcha = "reCaptcha"
	CaptchaTypeHCaptcha  = "hCaptcha"
)

// NewBrowserContext створює новий контекст chromedp з налаштуваннями проксі.
func NewBrowserContext(proxyAddress string) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),             // Вимикаємо безголовий режим для візуалізації
		chromedp.Flag("disable-gpu", true),          // Рекомендується для стабільності
		chromedp.Flag("no-sandbox", true),           // Рекомендується для Docker-оточення
		chromedp.Flag("ignore-certificate-errors", true), // Ігноруємо помилки сертифікатів
	)
	if proxyAddress != "" {
		log.Printf("Використовуємо проксі: %s", proxyAddress)
		opts = append(opts, chromedp.Flag("proxy-server", proxyAddress))
	}

	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancelCtx := chromedp.NewContext(
		allocatorCtx,
		chromedp.WithLogf(log.Printf), // Для логування дій chromedp
	)

	return ctx, func() {
		cancelCtx()
		cancelAllocator()
	}
}

// NavigateToURL переходить за вказаною URL-адресою в браузері.
func NavigateToURL(ctx context.Context, url string) error {
	log.Printf("Переходимо за URL: %s", url)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second), // Зачекаємо завантаження сторінки
	)
	if err != nil {
		log.Printf("Помилка навігації за URL %s: %v", url, err)
	} else {
		log.Printf("Успішно перейшли за URL: %s", url)
	}
	return err
}

// RefreshPage оновлює поточну сторінку в браузері.
func RefreshPage(ctx context.Context) error {
	log.Println("Оновлюємо сторінку...")
	err := chromedp.Run(ctx,
		chromedp.Reload(),
		chromedp.Sleep(3*time.Second), // Даємо час на оновлення
	)
	if err != nil {
		log.Printf("Не вдалося оновити сторінку: %v", err)
	} else {
		log.Println("Сторінку успішно оновлено.")
	}
	return err
}

// EnterETHAddress вводить ETH-адресу в поле вводу.
func EnterETHAddress(ctx context.Context, ethAddress string) error {
	placeholderText := "Please enter ETH address or ENS name" // Точний текст placeholder'а
	selector := fmt.Sprintf(`input[placeholder="%s"]`, placeholderText) // CSS-селектор

	log.Printf("Намагаємося ввести ETH-адресу у поле з placeholder'ом '%s'", placeholderText)
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery), // Чекаємо, доки поле стане видимим
		chromedp.Clear(selector, chromedp.ByQuery),       // Очищаємо поле
		chromedp.SendKeys(selector, ethAddress, chromedp.ByQuery), // Вводимо адресу
		chromedp.Sleep(1*time.Second), // Невелика затримка, щоб переконатися, що JS обробив введення
	)
	if err != nil {
		log.Printf("Не вдалося ввести ETH-адресу '%s' у поле з placeholder'ом '%s': %v", ethAddress, placeholderText, err)
	} else {
		log.Printf("Успішно введено ETH-адресу '%s' у поле з placeholder'ом '%s'.", ethAddress, placeholderText)
	}
	return err
}

// DetectCaptchaType визначає, який тип капчі (hCaptcha, reCAPTCHA або жодна) присутній на сторінці.
func DetectCaptchaType(ctx context.Context) (string, error) {
	log.Println("Визначаємо тип капчі...")

	var hCaptchaExists bool
	var reCaptchaExists bool

	// Максимальний час очікування для появи будь-якої капчі на сторінці
	const captchaDetectionTimeout = 15 * time.Second

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Створюємо контекст для операцій очікування з таймаутом
			waitCtx, cancelWait := context.WithTimeout(ctx, captchaDetectionTimeout)
			defer cancelWait()

			reCaptchaSel := `iframe[src*="google.com/recaptcha"]`
			hCaptchaSel := `iframe[src*="hcaptcha.com/"]`

			// Канал для сигналу, яка капча була знайдена першою
			done := make(chan string, 1)

			// Запускаємо паралельні горутини для очікування кожного типу капчі
			go func() {
				if err := chromedp.Run(waitCtx, chromedp.WaitVisible(reCaptchaSel, chromedp.ByQuery)); err == nil {
					select {
					case done <- CaptchaTypeRecaptcha:
					case <-waitCtx.Done(): // Перевіряємо, чи контекст вже скасовано
					}
				}
			}()

			go func() {
				if err := chromedp.Run(waitCtx, chromedp.WaitVisible(hCaptchaSel, chromedp.ByQuery)); err == nil {
					select {
					case done <- CaptchaTypeHCaptcha:
					case <-waitCtx.Done(): // Перевіряємо, чи контекст вже скасовано
					}
				}
			}()

			// Чекаємо, доки одна з капч буде виявлена, або закінчиться таймаут
			select {
			case captchaType := <-done:
				if captchaType == CaptchaTypeRecaptcha {
					reCaptchaExists = true
					log.Println("iframe reCAPTCHA став видимим.")
				} else if captchaType == CaptchaTypeHCaptcha {
					hCaptchaExists = true
					log.Println("iframe hCaptcha став видимим.")
				}
				return nil
			case <-waitCtx.Done():
				// Таймаут очікування, жоден iframe капчі не став видимим
				return fmt.Errorf("таймаут очікування, жоден iframe капчі не став видимим: %w", waitCtx.Err())
			}
		}),
	)

	if err != nil {
		// Якщо виник таймаут або інша помилка під час очікування, повертаємо її
		return CaptchaTypeNone, err
	}

	if reCaptchaExists {
		return CaptchaTypeRecaptcha, nil
	}
	if hCaptchaExists {
		return CaptchaTypeHCaptcha, nil
	}

	log.Println("Капчу не виявлено після очікування.")
	return CaptchaTypeNone, nil
}

// CheckForInvalidCaptchaError перевіряє наявність повідомлення про помилку INVALID_CAPTCHA на сторінці.
func CheckForInvalidCaptchaError(ctx context.Context) (bool, error) {
	log.Println("Перевіряємо наявність помилки [INVALID_CAPTCHA]...")
	var exists bool
	// Шукаємо елемент, який містить текст помилки. Це може бути span, div тощо.
	// Припускаємо, що помилка з'являється десь на сторінці як видимий текст.
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.innerText.includes("[INVALID_CAPTCHA]")`, &exists),
	)
	if err != nil {
		return false, fmt.Errorf("помилка під час перевірки повідомлення про помилку капчі: %w", err)
	}
	if exists {
		log.Println("Виявлено помилку: [INVALID_CAPTCHA].")
	} else {
		log.Println("Помилку [INVALID_CAPTCHA] не виявлено.")
	}
	return exists, nil
}

// WaitForMiningAndClaim чекає, поки на сторінці з'явиться кнопка "Claim reward", і натискає її.
// Параметр durationHours вказує, скільки годин чекати до натискання.
func WaitForMiningAndClaim(ctx context.Context, durationHours time.Duration) error {
	log.Printf("Чекаємо %v годин до натискання кнопки 'Claim reward'...", durationHours)
	time.Sleep(durationHours * time.Hour) // Чекаємо задану кількість годин

	claimButtonSelector := `button.claim-reward` // Припускаємо такий селектор для кнопки "Claim reward"

	log.Println("Намагаємося натиснути кнопку 'Claim reward'...")
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(claimButtonSelector, chromedp.ByQuery), // Чекаємо, доки кнопка стане видимою
		chromedp.Click(claimButtonSelector, chromedp.ByQuery),       // Натискаємо кнопку
		chromedp.Sleep(5*time.Second), // Даємо час на обробку кліку
	)
	if err != nil {
		log.Printf("Не вдалося натиснути кнопку 'Claim reward': %v", err)
		return fmt.Errorf("не вдалося натиснути кнопку 'Claim reward': %w", err)
	}
	log.Println("Кнопку 'Claim reward' успішно натиснуто.")
	return nil
}

func InjectCaptchaToken(ctx context.Context, token string) error {
	js := fmt.Sprintf(`
		(() => {
			document.querySelector('textarea[name="h-captcha-response"]').value = "%s";
			window.hcaptcha = window.hcaptcha || {};
			window.hcaptcha.getResponse = () => "%s";
		})();
	`, token, token)

	return chromedp.Run(ctx,
		chromedp.Evaluate(js, nil),
	)
}

func ClickStartMining(ctx context.Context) error {
	return chromedp.Run(ctx,
		chromedp.Click("button", chromedp.NodeVisible),
	)
}
