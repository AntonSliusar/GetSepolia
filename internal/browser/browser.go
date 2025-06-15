package browser

import (
	"context"
	"log"

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
