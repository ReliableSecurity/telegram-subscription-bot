package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CryptoUtils struct {
	client *http.Client
}

type CoinGeckoResponse struct {
	USD float64 `json:"usd"`
	EUR float64 `json:"eur"`
	RUB float64 `json:"rub"`
}

type CoinGeckoPriceResponse map[string]CoinGeckoResponse

func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *CryptoUtils) ConvertToCrypto(amount float64, fromCurrency, toCurrency string) (float64, error) {
	rate, err := c.getCryptoRate(toCurrency, fromCurrency)
	if err != nil {
		return 0, err
	}
	
	return amount / rate, nil
}

func (c *CryptoUtils) getCryptoRate(cryptoCurrency, fiatCurrency string) (float64, error) {
	cryptoID := c.getCryptoID(cryptoCurrency)
	if cryptoID == "" {
		return 0, fmt.Errorf("unsupported cryptocurrency: %s", cryptoCurrency)
	}
	
	fiatLower := strings.ToLower(fiatCurrency)
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", cryptoID, fiatLower)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	var priceResponse CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		return 0, err
	}
	
	cryptoData, exists := priceResponse[cryptoID]
	if !exists {
		return 0, fmt.Errorf("price not found for %s", cryptoCurrency)
	}
	
	switch fiatLower {
	case "usd":
		return cryptoData.USD, nil
	case "eur":
		return cryptoData.EUR, nil
	case "rub":
		return cryptoData.RUB, nil
	default:
		return 0, fmt.Errorf("unsupported fiat currency: %s", fiatCurrency)
	}
}

func (c *CryptoUtils) getCryptoID(currency string) string {
	switch strings.ToUpper(currency) {
	case "BTC":
		return "bitcoin"
	case "ETH":
		return "ethereum"
	case "USDT":
		return "tether"
	case "LTC":
		return "litecoin"
	case "BCH":
		return "bitcoin-cash"
	case "XRP":
		return "ripple"
	case "ADA":
		return "cardano"
	case "DOT":
		return "polkadot"
	case "BNB":
		return "binancecoin"
	case "LINK":
		return "chainlink"
	default:
		return ""
	}
}

func (c *CryptoUtils) CheckPayment(address, expectedAmount, currency string) (bool, error) {
	// This is a simplified implementation
	// In a real application, you would integrate with blockchain APIs
	
	switch strings.ToUpper(currency) {
	case "BTC":
		return c.checkBTCPayment(address, expectedAmount)
	case "ETH":
		return c.checkETHPayment(address, expectedAmount)
	case "USDT":
		return c.checkUSDTPayment(address, expectedAmount)
	default:
		return false, fmt.Errorf("unsupported cryptocurrency: %s", currency)
	}
}

func (c *CryptoUtils) checkBTCPayment(address, expectedAmount string) (bool, error) {
	// Integrate with BlockCypher, Blockchain.info, or similar API
	// This is a placeholder implementation
	
	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/main/addrs/%s", address)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	
	var addressInfo map[string]interface{}
	if err := json.Unmarshal(body, &addressInfo); err != nil {
		return false, err
	}
	
	// Check if the address has received the expected amount
	if balance, exists := addressInfo["balance"]; exists {
		balanceFloat, _ := strconv.ParseFloat(fmt.Sprintf("%v", balance), 64)
		expectedFloat, _ := strconv.ParseFloat(expectedAmount, 64)
		
		// Convert satoshis to BTC
		balanceBTC := balanceFloat / 100000000
		
		return balanceBTC >= expectedFloat, nil
	}
	
	return false, nil
}

func (c *CryptoUtils) checkETHPayment(address, expectedAmount string) (bool, error) {
	// Integrate with Etherscan, Infura, or similar API
	// This is a placeholder implementation
	
	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest", address)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	
	var ethResponse map[string]interface{}
	if err := json.Unmarshal(body, &ethResponse); err != nil {
		return false, err
	}
	
	if result, exists := ethResponse["result"]; exists {
		balanceWei, _ := strconv.ParseFloat(fmt.Sprintf("%v", result), 64)
		expectedFloat, _ := strconv.ParseFloat(expectedAmount, 64)
		
		// Convert wei to ETH
		balanceETH := balanceWei / 1000000000000000000
		
		return balanceETH >= expectedFloat, nil
	}
	
	return false, nil
}

func (c *CryptoUtils) checkUSDTPayment(address, expectedAmount string) (bool, error) {
	// USDT can be on different networks (Ethereum, Tron, etc.)
	// This would need to check the appropriate network
	// For now, we'll check USDT on Ethereum network
	
	return c.checkETHPayment(address, expectedAmount)
}

func (c *CryptoUtils) GeneratePaymentAddress(currency string) (string, error) {
	// This would integrate with a wallet service or generate addresses
	// For now, return placeholder addresses
	
	switch strings.ToUpper(currency) {
	case "BTC":
		return "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", nil
	case "ETH":
		return "0x0000000000000000000000000000000000000000", nil
	case "USDT":
		return "0x0000000000000000000000000000000000000000", nil
	default:
		return "", fmt.Errorf("unsupported cryptocurrency: %s", currency)
	}
}

func (c *CryptoUtils) GetSupportedCurrencies() []string {
	return []string{"BTC", "ETH", "USDT", "LTC", "BCH", "XRP", "ADA", "DOT", "BNB", "LINK"}
}

func (c *CryptoUtils) FormatCryptoAmount(amount float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "BTC":
		return fmt.Sprintf("%.8f", amount)
	case "ETH":
		return fmt.Sprintf("%.6f", amount)
	case "USDT":
		return fmt.Sprintf("%.2f", amount)
	default:
		return fmt.Sprintf("%.6f", amount)
	}
}
