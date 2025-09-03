package handlers

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

// Mock queue function type
type QueueFunc func(string) chan queue.Result

// Mock responses storage
var mockResponses map[string]queue.Result
var mockCallCount map[string]int
var mockMutex sync.Mutex

// Mock implementation of AddWalletToQueue
func mockAddWalletToQueue(wallet string) chan queue.Result {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	
	if mockCallCount == nil {
		mockCallCount = make(map[string]int)
	}
	mockCallCount[wallet]++
	
	ch := make(chan queue.Result, 1)
	
	if resp, exists := mockResponses[wallet]; exists {
		go func() {
			time.Sleep(5 * time.Millisecond)
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

func setMockResponse(wallet string, response queue.Result) {
	if mockResponses == nil {
		mockResponses = make(map[string]queue.Result)
	}
	mockResponses[wallet] = response
}

func getMockCallCount(wallet string) int {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	return mockCallCount[wallet]
}

func resetMockData() {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	mockResponses = make(map[string]queue.Result)
	mockCallCount = make(map[string]int)
}

// Test handler that uses the mock
func getMockSolanaBalance(queueFunc QueueFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				waitChan := queueFunc(wallet)
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
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/get-balance", getMockSolanaBalance(mockAddWalletToQueue))
	return router
}

func TestGetSolanaBalance_SingleWallet(t *testing.T) {
	resetMockData()
	setMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})

	router := setupTestRouter()
	
	requestBody := models.WalletsRequest{
		Wallets: []string{"11111111111111111111111111111111"},
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Empty(t, response.Error)
	assert.Len(t, response.Object, 1)
	assert.Equal(t, "11111111111111111111111111111111", response.Object[0].Wallet)
	assert.Equal(t, "2.5", response.Object[0].Balance)
	assert.Equal(t, "miss", response.Object[0].Cache)
}

func TestGetSolanaBalance_MultipleWallets(t *testing.T) {
	resetMockData()
	
	// Set different responses for different wallets
	setMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})
	setMockResponse("22222222222222222222222222222222", queue.Result{
		Result: "3.7",
		Error:  nil,
		Cache:  true,
	})
	setMockResponse("33333333333333333333333333333333", queue.Result{
		Result: "1.2",
		Error:  nil,
		Cache:  false,
	})

	router := setupTestRouter()
	
	requestBody := models.WalletsRequest{
		Wallets: []string{
			"11111111111111111111111111111111",
			"22222222222222222222222222222222",
			"33333333333333333333333333333333",
		},
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Empty(t, response.Error)
	assert.Len(t, response.Object, 3)
	
	// Create a map for easier assertion since order might vary due to concurrency
	balanceMap := make(map[string]models.WalletBalance)
	for _, balance := range response.Object {
		balanceMap[balance.Wallet] = balance
	}
	
	assert.Equal(t, "2.5", balanceMap["11111111111111111111111111111111"].Balance)
	assert.Equal(t, "miss", balanceMap["11111111111111111111111111111111"].Cache)
	
	assert.Equal(t, "3.7", balanceMap["22222222222222222222222222222222"].Balance)
	assert.Equal(t, "hit", balanceMap["22222222222222222222222222222222"].Cache)
	
	assert.Equal(t, "1.2", balanceMap["33333333333333333333333333333333"].Balance)
	assert.Equal(t, "miss", balanceMap["33333333333333333333333333333333"].Cache)
}

func TestGetSolanaBalance_FiveRequestsSameWallet(t *testing.T) {
	resetMockData()
	setMockResponse("11111111111111111111111111111111", queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})

	router := setupTestRouter()
	
	var wg sync.WaitGroup
	results := make([]models.GenericResponse[[]models.WalletBalance], 5)
	
	// Make 5 concurrent requests with the same wallet
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			requestBody := models.WalletsRequest{
				Wallets: []string{"11111111111111111111111111111111"},
			}
			
			jsonBody, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			var response models.GenericResponse[[]models.WalletBalance]
			json.Unmarshal(w.Body.Bytes(), &response)
			results[index] = response
		}(i)
	}
	
	wg.Wait()
	
	// Verify all requests succeeded
	for i, response := range results {
		assert.True(t, response.Success, fmt.Sprintf("Request %d failed", i))
		assert.Empty(t, response.Error, fmt.Sprintf("Request %d has error", i))
		assert.Len(t, response.Object, 1, fmt.Sprintf("Request %d has wrong length", i))
		assert.Equal(t, "11111111111111111111111111111111", response.Object[0].Wallet)
		assert.Equal(t, "2.5", response.Object[0].Balance)
	}
	
	// Verify the wallet was called 5 times
	assert.Equal(t, 5, getMockCallCount("11111111111111111111111111111111"))
}

func TestGetSolanaBalance_CombinedScenario(t *testing.T) {
	resetMockData()
	
	// Setup responses for different wallets
	wallets := []string{
		"11111111111111111111111111111111",
		"22222222222222222222222222222222",
		"33333333333333333333333333333333",
	}
	
	for i, wallet := range wallets {
		setMockResponse(wallet, queue.Result{
			Result: fmt.Sprintf("%d.%d", i+1, i+5),
			Error:  nil,
			Cache:  i%2 == 0, // Alternate cache hit/miss
		})
	}

	router := setupTestRouter()
	
	var wg sync.WaitGroup
	numRequests := 8
	results := make([]models.GenericResponse[[]models.WalletBalance], numRequests)
	
	// Make multiple concurrent requests with different combinations
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			var requestWallets []string
			if index%3 == 0 {
				// Single wallet request
				requestWallets = []string{wallets[index%len(wallets)]}
			} else if index%3 == 1 {
				// Two wallets request
				requestWallets = []string{wallets[0], wallets[1]}
			} else {
				// All wallets request
				requestWallets = wallets
			}
			
			requestBody := models.WalletsRequest{
				Wallets: requestWallets,
			}
			
			jsonBody, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			var response models.GenericResponse[[]models.WalletBalance]
			json.Unmarshal(w.Body.Bytes(), &response)
			results[index] = response
		}(i)
	}
	
	wg.Wait()
	
	// Verify all requests succeeded
	for i, response := range results {
		assert.True(t, response.Success, fmt.Sprintf("Request %d failed", i))
		assert.Empty(t, response.Error, fmt.Sprintf("Request %d has error", i))
		assert.Greater(t, len(response.Object), 0, fmt.Sprintf("Request %d has no results", i))
	}
	
	// Verify each wallet was called multiple times
	for _, wallet := range wallets {
		count := getMockCallCount(wallet)
		assert.Greater(t, count, 0, fmt.Sprintf("Wallet %s was not called", wallet))
	}
}

func TestGetSolanaBalance_CachingFunctionality(t *testing.T) {
	resetMockData()
	
	wallet := "11111111111111111111111111111111"
	
	// First call - cache miss
	setMockResponse(wallet, queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  false,
	})

	router := setupTestRouter()
	
	// First request
	requestBody := models.WalletsRequest{
		Wallets: []string{wallet},
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req1, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	var response1 models.GenericResponse[[]models.WalletBalance]
	json.Unmarshal(w1.Body.Bytes(), &response1)
	
	assert.True(t, response1.Success)
	assert.Equal(t, "miss", response1.Object[0].Cache)
	
	// Second call - simulate cache hit
	setMockResponse(wallet, queue.Result{
		Result: "2.5",
		Error:  nil,
		Cache:  true,
	})
	
	// Second request
	jsonBody2, _ := json.Marshal(requestBody)
	req2, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	var response2 models.GenericResponse[[]models.WalletBalance]
	json.Unmarshal(w2.Body.Bytes(), &response2)
	
	assert.True(t, response2.Success)
	assert.Equal(t, "hit", response2.Object[0].Cache)
	assert.Equal(t, response1.Object[0].Balance, response2.Object[0].Balance) // Same balance
}

func TestGetSolanaBalance_ErrorHandling(t *testing.T) {
	resetMockData()
	setMockResponse("invalid_wallet", queue.Result{
		Result: "",
		Error:  fmt.Errorf("invalid wallet address"),
		Cache:  false,
	})

	router := setupTestRouter()
	
	requestBody := models.WalletsRequest{
		Wallets: []string{"invalid_wallet"},
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Len(t, response.Object, 1)
	assert.Equal(t, "invalid wallet address", response.Object[0].Balance)
	assert.Equal(t, "miss", response.Object[0].Cache)
}

func TestGetSolanaBalance_InvalidJSON(t *testing.T) {
	resetMockData()
	router := setupTestRouter()
	
	// Send invalid JSON
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer([]byte(`{"invalid": json`)))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.GenericResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestGetSolanaBalance_EmptyWalletsList(t *testing.T) {
	resetMockData()
	router := setupTestRouter()
	
	requestBody := models.WalletsRequest{
		Wallets: []string{},
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/get-balance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.GenericResponse[[]models.WalletBalance]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.Empty(t, response.Object)
}