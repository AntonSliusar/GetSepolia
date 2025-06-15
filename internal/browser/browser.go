package browser

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	// Using runner for a more basic setup
)


func NewBrowserContext(proxyAddress string) (context.Context, context.CancelFunc) {
	
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),                  // Вимикаємо безголовий режим
		chromedp.Flag("disable-gpu", true),                // Рекомендується
		chromedp.Flag("no-sandbox", true),                 // Рекомендується
		chromedp.Flag("ignore-certificate-errors", true),
	)
	if proxyAddress != "" {
		log.Printf("Using proxy: %s", proxyAddress)
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


func NavigateToURL(ctx context.Context, url string) error {
	log.Printf("Navigating to URL: %s", url)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
	)
	if err != nil {
		log.Printf("Error navigating to URL %s: %v", url, err)
	} else {
		log.Printf("Successfully navigated to URL: %s", url)
	}
	return err
}

func EnterETHAddress(ctx context.Context, ethAddress string) error {
	time.Sleep(2 * time.Second) 

	placeholderText := "Please enter ETH address or ENS name" // Точний текст placeholder'а
	selector := fmt.Sprintf(`input[placeholder="%s"]`, placeholderText) // CSS-селектор

	log.Printf("Attempting to enter ETH address into field with placeholder '%s'", placeholderText)
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery), // Чекаємо, доки поле стане видимим
		chromedp.Clear(selector, chromedp.ByQuery),      // Очищаємо поле
		chromedp.SendKeys(selector, ethAddress, chromedp.ByQuery), // Вводимо адресу
		chromedp.Sleep(1*time.Second), // Невелика затримка, щоб переконатися, що JS обробив введення
	)
	if err != nil {
		log.Printf("Failed to enter ETH address '%s' into field with placeholder '%s': %v", ethAddress, placeholderText, err)
	} else {
		log.Printf("Successfully entered ETH address '%s' into field with placeholder '%s'.", ethAddress, placeholderText)
	}
	return err
}

func SetRecaptchaResponse(ctx context.Context, token string) error {
	log.Println("Setting reCAPTCHA response token...")
	selector := `textarea[name="g-recaptcha-response"]`

	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
            var el = document.querySelector('%s');
            if (el) {
                el.value = '%s';
                el.dispatchEvent(new Event('change', { bubbles: true })); // Імітуємо подію change
                el.dispatchEvent(new Event('input', { bubbles: true }));  // Імітуємо подію input
                console.log('reCAPTCHA token set and events dispatched.'); // Для дебагу в консолі браузера
            } else {
                console.error('reCAPTCHA textarea not found by selector: %s');
            }
        `, selector, token, selector), nil),
		chromedp.Sleep(2*time.Second), // Збільшимо затримку трохи
	)
	if err != nil {
		log.Printf("Failed to set reCAPTCHA response token: %v", err)
	} else {
		log.Println("Successfully set reCAPTCHA response token and dispatched events.")
	}
	return err
}

func ClickStartMiningButton(ctx context.Context) error {
	selector := `button.start-action` // CSS-селектор для кнопки

	log.Println("Attempting to click 'Start Mining' button...")
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery), // Чекаємо, доки кнопка стане видимою
		chromedp.Click(selector, chromedp.ByQuery),       // Натискаємо кнопку
		chromedp.Sleep(3*time.Second),                    // Даємо час на обробку кліку та можливе завантаження
	)
	if err != nil {
		log.Printf("Failed to click 'Start Mining' button: %v", err)
	} else {
		log.Println("Successfully clicked 'Start Mining' button.")
	}
	return err
}






























































///----------------------------------------------------------------------------



// // NewBrowserContext створює новий контекст chromedp для взаємодії з браузером.
// // Він налаштовує віддалений підключення до Chrome, що працює в Docker.
// func NewBrowserContext() (context.Context, context.CancelFunc) {
// 	// Отримуємо URL до Chrome DevTools Protocol від змінної середовища.
// 	// Це очікується від налаштувань Docker.
// 	chromeWsURL := os.Getenv("CHROME_WS_URL")
// 	if chromeWsURL == "" {
// 		// Якщо змінна не встановлена, припускаємо, що Chrome працює локально або за замовчуванням Docker IP
// 		// Для локальної розробки, можна використовувати:
// 		// chromeWsURL = "ws://127.0.0.1:9222"
// 		// Для Docker, зазвичай, це буде IP контейнера:9222
// 		log.Fatal("CHROME_WS_URL environment variable is not set. Please provide the WebSocket URL for the Chrome instance.")
// 	}

// 	allocatorCtx, cancelAllocator := chromedp.NewRemoteAllocator(context.Background(), chromeWsURL)

// 	// Створюємо контекст chromedp
// 	// Без headless, щоб можна було бачити браузер для дебагінгу
// 	// Додаємо опції для ігнорування сертифікатів SSL та збільшення таймауту, якщо потрібно
// 	ctx, cancelCtx := chromedp.NewContext(
// 		allocatorCtx,
// 		chromedp.WithLogf(log.Printf), // Для логування дій chromedp
// 	)

// 	// Повертаємо контекст та функцію скасування для коректного закриття
// 	return ctx, func() {
// 		cancelCtx()
// 		cancelAllocator()
// 	}
// }

// // NavigateToURL переходить за вказаною URL-адресою в браузері.
