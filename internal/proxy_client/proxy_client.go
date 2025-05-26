package proxyclient

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func CreateClientWirtProxy(proxyURL string) (*http.Client, error) {
	parsedProxyURL, err := url.Parse(proxyURL)        /// Парсимо проксі 
	if err != nil {
		return nil, fmt.Errorf("помилка парсингу URL проксі '%s': %w", proxyURL, err)   /// Додати логер
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout: 30 * time.Second,
	}

	return client, nil

}