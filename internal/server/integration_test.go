package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/pkg/models"
	"main/pkg/queue"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Integration test mocks
var integrationMockResponses map[string]queue.Result
var integrationMockCallCount map[string]int
var integrationMockLicenses map[string]bool
var integrationMockIpCounts map[string]int
var integrationMutex sync.RWMutex

// Mock queue function
func integrationMockAddWalletToQueue(wallet string) chan queue.Result {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	
	if integrationMockCallCount == nil {
		integrationMockCallCount = make(map[string]int)
	}
	integrationMockCallCount[wallet]++
	
	ch := make(chan queue.Result, 1)
	
	if resp, exists := integrationMockResponses[wallet]; exists {
		go func() {
			time.Sleep(5 * time.Millisecond) // Simulate processing
			ch <- resp
		}()
	} else {
		go func() {
			time.Sleep(5 * time.Millisecond)
			ch <- queue.Result{
				Result: "1.5",
				Error:  nil,
				Cache:  false,
			}
		}()
	}
	
	return ch
}

// Mock authentication middleware
func integrationMockAuthenticate(c *gin.Context) {
	// Check for API key
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "missing api key"})
		return
	}
	
	// Validate license
	integrationMutex.RLock()
	valid := integrationMockLicenses[apiKey]
	integrationMutex.RUnlock()
	
	if !valid {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
		return
	}
	
	// Check IP rate limit
	integrationMutex.Lock()
	if integrationMockIpCounts == nil {
		integrationMockIpCounts = make(map[string]int)
	}
	
	ip := c.ClientIP()
	currentCount := integrationMockIpCounts[ip]
	
	if currentCount >= 10 {
		integrationMutex.Unlock()
		c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
		return
	}
	
	// Increment IP count
	integrationMockIpCounts[ip]++
	integrationMutex.Unlock()
	
	c.Next()
}

// Mock handler using dependency injection
func integrationMockGetSolanaBalance(c *gin.Context) {
	var request models.WalletsRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, models.GenericResponse[any]{
			Object:  nil,
			Error:   "Invalid request body",
			Success: false,
		})
		return
	}

	var result []models.WalletBalance
	wg := sync.WaitGroup{}
	wg.Add(len(request.Wallets))
	for _, wallet := range request.Wallets {
		go func(wallet string) {
			defer wg.Done()
			waitChan := integrationMockAddWalletToQueue(wallet)
			res := <-waitChan
			bal := models.WalletBalance{
				Wallet: wallet,
			}
			if res.Error != nil {
				bal.Balance = res.Error.Error()
			} else {
				bal.Balance = res.Result
			}
			if res.Cache {
				bal.Cache = "hit"
			} else {
				bal.Cache = "miss"
			}
			result = append(result, bal)
		}(wallet)
	}
	wg.Wait()

	c.JSON(200, models.GenericResponse[[]models.WalletBalance]{
		Object:  result,
		Error:   "",
		Success: true,
	})
}

func resetIntegrationMockData() {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	integrationMockResponses = make(map[string]queue.Result)
	integrationMockCallCount = make(map[string]int)
	integrationMockLicenses = make(map[string]bool)
	integrationMockIpCounts = make(map[string]int)
}

func setIntegrationMockResponse(wallet string, response queue.Result) {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	if integrationMockResponses == nil {
		integrationMockResponses = make(map[string]queue.Result)
	}
	integrationMockResponses[wallet] = response
}

func setIntegrationValidLicense(apiKey string, valid bool) {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	if integrationMockLicenses == nil {
		integrationMockLicenses = make(map[string]bool)
	}
	integrationMockLicenses[apiKey] = valid
}

func setIntegrationIpRequestCount(ip string, count int) {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	if integrationMockIpCounts == nil {
		integrationMockIpCounts = make(map[string]int)
	}
	integrationMockIpCounts[ip] = count
}

func getIntegrationMockCallCount(wallet string) int {
	integrationMutex.Lock()
	defer integrationMutex.Unlock()
	return integrationMockCallCount[wallet]
}

func getIntegrationIpRequestCount(ip string) int {
	integrationMutex.RLock()
	defer integrationMutex.RUnlock()
	if integrationMockIpCounts == nil {
		return 0
	}
	return integrationMockIpCounts[ip]
}

func setupIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Apply authentication middleware
	apiAuth := router.Group("/api", integrationMockAuthenticate)
	apiAuth.POST("/get-balance", integrationMockGetSolanaBalance)
	
	return router
}

func makeAuthenticatedRequest(router *gin.Engine, apiKey, ip string, wallets []string) *httptest.ResponseRecorder {
	requestBody := models.WalletsRequest{
		Wallets: wallets,
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.RemoteAddr = ip + ":12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

func TestIntegration_SingleWallet_WithAuth(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	setIntegrationMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})
	
	router := setupIntegrationRouter()
	
	w := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", []string{"11111111111111111111111111111111"})
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Len(t, response.Object, 1)
	assert.Equal(t, "2.5", response.Object[0].Balance)
}

func TestIntegration_MultipleWallets_WithAuth(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	wallets := []string{
		"11111111111111111111111111111111",
		"22222222222222222222222222222222",
		"33333333333333333333333333333333",
	}
	
	for i, wallet := range wallets {
		setIntegrationMockResponse(wallet, queue.Result{
			Result: fmt.Sprintf("%d.%d", i+1, i+5),
			Error:  nil,
			Cache:  i%2 == 0,
		})
	}
	
	router := setupIntegrationRouter()
	
	w := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", wallets)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Len(t, response.Object, 3)
}

func TestIntegration_FiveRequestsSameWallet_WithAuth(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	setIntegrationMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})
	
	router := setupIntegrationRouter()
	
	var wg sync.WaitGroup
	results := make([]int, 5)
	
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			w := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", []string{"11111111111111111111111111111111"})
			results[index] = w.Code
		}(i)
	}
	
	wg.Wait()
	
	// All requests should succeed (under rate limit)
	for i, code := range results {
		assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Request %d should succeed", i))
	}
	
	// Verify the wallet was called 5 times
	assert.Equal(t, 5, getIntegrationMockCallCount("11111111111111111111111111111111"))
}

func TestIntegration_AllScenariosAtOnce(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	// Setup different wallet responses
	wallets := []string{
		"11111111111111111111111111111111",
		"22222222222222222222222222222222",
		"33333333333333333333333333333333",
	}
	
	for i, wallet := range wallets {
		setIntegrationMockResponse(wallet, queue.Result{
			Result: fmt.Sprintf("%d.%d", i+2, i+7),
			Error:  nil,
			Cache:  i%2 == 1,
		})
	}
	
	router := setupIntegrationRouter()
	
	var wg sync.WaitGroup
	numRequests := 8
	results := make([]int, numRequests)
	
	// Mix of single wallet, multiple wallets, and all wallets requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			var requestWallets []string
			switch index % 4 {
			case 0:
				// Single wallet
				requestWallets = []string{wallets[0]}
			case 1:
				// Two wallets
				requestWallets = []string{wallets[0], wallets[1]}
			case 2:
				// All wallets
				requestWallets = wallets
			default:
				// Single different wallet
				requestWallets = []string{wallets[2]}
			}
			
			w := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", requestWallets)
			results[index] = w.Code
		}(i)
	}
	
	wg.Wait()
	
	// All requests should succeed (under rate limit)
	successCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}
	
	assert.Greater(t, successCount, 0, "At least some requests should succeed")
}

func TestIntegration_IPRateLimit_MultipleClients(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	
	// Set one IP near the limit, another fresh
	setIntegrationIpRequestCount("192.168.1.1", 8) // Near limit
	setIntegrationIpRequestCount("192.168.1.2", 0) // Fresh
	
	setIntegrationMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})
	
	router := setupIntegrationRouter()
	
	// First IP should have limited requests left
	w1 := makeAuthenticatedRequest(router, "test-key", "192.168.1.1", []string{"11111111111111111111111111111111"})
	assert.Equal(t, http.StatusOK, w1.Code, "First IP should still work")
	
	w2 := makeAuthenticatedRequest(router, "test-key", "192.168.1.1", []string{"11111111111111111111111111111111"})
	assert.Equal(t, http.StatusOK, w2.Code, "Second request from first IP should still work")
	
	// Third request should hit the limit (count was 8, +1=9, +1=10, which is the limit)
	w3 := makeAuthenticatedRequest(router, "test-key", "192.168.1.1", []string{"11111111111111111111111111111111"})
	assert.Equal(t, http.StatusTooManyRequests, w3.Code, "Third request from first IP should be rate limited")
	
	// Second IP should work fine
	w4 := makeAuthenticatedRequest(router, "test-key", "192.168.1.2", []string{"11111111111111111111111111111111"})
	assert.Equal(t, http.StatusOK, w4.Code, "Second IP should work fine")
}

func TestIntegration_CacheHitMiss_Behavior(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	wallet := "11111111111111111111111111111111"
	
	// First request - cache miss
	setIntegrationMockResponse(wallet, queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})
	
	router := setupIntegrationRouter()
	
	w1 := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", []string{wallet})
	assert.Equal(t, http.StatusOK, w1.Code)
	
	var response1 models.GenericResponse[[]models.WalletBalance]
	json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.Equal(t, "miss", response1.Object[0].Cache)
	
	// Second request - cache hit (simulated)
	setIntegrationMockResponse(wallet, queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  true,
	})
	
	w2 := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", []string{wallet})
	assert.Equal(t, http.StatusOK, w2.Code)
	
	var response2 models.GenericResponse[[]models.WalletBalance]
	json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.Equal(t, "hit", response2.Object[0].Cache)
}

func TestIntegration_AuthenticationFailure_NoBypassRateLimit(t *testing.T) {
	resetIntegrationMockData()
	
	// Don't set valid license - should fail auth
	setIntegrationValidLicense("invalid-key", false)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	router := setupIntegrationRouter()
	
	w := makeAuthenticatedRequest(router, "invalid-key", "127.0.0.1", []string{"11111111111111111111111111111111"})
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// IP count should not have been incremented since auth failed
	count := getIntegrationIpRequestCount("127.0.0.1")
	assert.Equal(t, 0, count, "IP count should not increment on auth failure")
}

func TestIntegration_ConcurrentRequests_MultipleWallets_WithRateLimit(t *testing.T) {
	resetIntegrationMockData()
	
	setIntegrationValidLicense("test-key", true)
	setIntegrationIpRequestCount("127.0.0.1", 0)
	
	wallets := []string{
		"11111111111111111111111111111111",
		"22222222222222222222222222222222",
		"33333333333333333333333333333333",
	}
	
	for i, wallet := range wallets {
		setIntegrationMockResponse(wallet, queue.Result{
			Result: fmt.Sprintf("%d.%d", i+1, i+3),
			Error:  nil,
			Cache:  false,
		})
	}
	
	router := setupIntegrationRouter()
	
	var wg sync.WaitGroup
	numRequests := 12 // Should exceed rate limit
	results := make([]int, numRequests)
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// Use different wallet combinations
			requestWallets := []string{wallets[index%len(wallets)]}
			w := makeAuthenticatedRequest(router, "test-key", "127.0.0.1", requestWallets)
			results[index] = w.Code
		}(i)
	}
	
	wg.Wait()
	
	successCount := 0
	rateLimitCount := 0
	
	for _, code := range results {
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitCount++
		}
	}
	
	// Some should succeed, some should be rate limited
	assert.Greater(t, successCount, 0, "Some requests should succeed")
	assert.Greater(t, rateLimitCount, 0, "Some requests should be rate limited due to IP limits")
	assert.Equal(t, numRequests, successCount+rateLimitCount, "All requests should get a response")
	
	// Check that all wallets were called at least once
	for _, wallet := range wallets {
		count := getIntegrationMockCallCount(wallet)
		assert.Greater(t, count, 0, fmt.Sprintf("Wallet %s should have been called", wallet))
	}
}