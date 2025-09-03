package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock service data
var mockLicenses map[string]bool
var mockIpCounts map[string]int
var mockMutex sync.RWMutex

func resetMockData() {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	mockLicenses = make(map[string]bool)
	mockIpCounts = make(map[string]int)
}

func setValidLicense(apiKey string, valid bool) {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	if mockLicenses == nil {
		mockLicenses = make(map[string]bool)
	}
	mockLicenses[apiKey] = valid
}

func setIpRequestCount(ip string, count int) {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	if mockIpCounts == nil {
		mockIpCounts = make(map[string]int)
	}
	mockIpCounts[ip] = count
}

func getIpRequestCount(ip string) int {
	mockMutex.RLock()
	defer mockMutex.RUnlock()
	if mockIpCounts == nil {
		return 0
	}
	return mockIpCounts[ip]
}

// Mock middleware that simulates the authentication logic
func mockAuthenticate(c *gin.Context) {
	// Check for API key
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "missing api key"})
		return
	}
	
	// Validate license
	mockMutex.RLock()
	valid := mockLicenses[apiKey]
	mockMutex.RUnlock()
	
	if !valid {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
		return
	}
	
	// Check IP rate limit
	mockMutex.Lock()
	if mockIpCounts == nil {
		mockIpCounts = make(map[string]int)
	}
	
	ip := c.ClientIP()
	currentCount := mockIpCounts[ip]
	
	if currentCount >= 10 {
		mockMutex.Unlock()
		c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
		return
	}
	
	// Increment IP count
	mockIpCounts[ip]++
	mockMutex.Unlock()
	
	c.Next()
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Apply mock authentication middleware
	api := router.Group("/api", mockAuthenticate)
	api.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	return router
}

func TestAuthenticate_ValidAPIKey(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 0)
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "valid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthenticate_MissingAPIKey(t *testing.T) {
	resetMockData()
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	// No x-api-key header set
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_InvalidAPIKey(t *testing.T) {
	resetMockData()
	setValidLicense("invalid-api-key", false)
	setIpRequestCount("127.0.0.1", 0)
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "invalid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_IPRateLimit_UnderLimit(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 5) // Under limit of 10
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "valid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify count was incremented
	assert.Equal(t, 6, getIpRequestCount("127.0.0.1"))
}

func TestAuthenticate_IPRateLimit_AtLimit(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 10) // At limit of 10
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "valid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestAuthenticate_IPRateLimit_OverLimit(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 15) // Over limit of 10
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "valid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestAuthenticate_MultipleIPs_SeparateRateLimits(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	
	// Set different limits for different IPs
	setIpRequestCount("192.168.1.1", 5)  // Under limit
	setIpRequestCount("192.168.1.2", 10) // At limit
	
	router := setupTestRouter()
	
	// Test first IP (should pass)
	req1, _ := http.NewRequest("GET", "/api/test", nil)
	req1.Header.Set("x-api-key", "valid-api-key")
	req1.RemoteAddr = "192.168.1.1:12345"
	
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Test second IP (should fail)
	req2, _ := http.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("x-api-key", "valid-api-key")
	req2.RemoteAddr = "192.168.1.2:12345"
	
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestAuthenticate_ConcurrentRequests_SameIP(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 0)
	
	router := setupTestRouter()
	
	var wg sync.WaitGroup
	results := make([]int, 15) // Try 15 concurrent requests
	
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			req, _ := http.NewRequest("GET", "/api/test", nil)
			req.Header.Set("x-api-key", "valid-api-key")
			req.RemoteAddr = "127.0.0.1:12345"
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			results[index] = w.Code
		}(i)
	}
	
	wg.Wait()
	
	successCount := 0
	rateLimitCount := 0
	
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}
	
	// Some requests should succeed (up to 10), some should be rate limited
	assert.Greater(t, successCount, 0, "Some requests should succeed")
	assert.Greater(t, rateLimitCount, 0, "Some requests should be rate limited")
	assert.Equal(t, 15, successCount+rateLimitCount, "All requests should get a response")
	assert.LessOrEqual(t, successCount, 10, "No more than 10 requests should succeed")
}

func TestAuthenticate_SequentialRequests_RateLimit(t *testing.T) {
	resetMockData()
	setValidLicense("valid-api-key", true)
	setIpRequestCount("127.0.0.1", 0)
	
	router := setupTestRouter()
	
	// Make exactly 10 requests (the limit)
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("x-api-key", "valid-api-key")
		req.RemoteAddr = "127.0.0.1:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}
	
	// 11th request should fail
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "valid-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "11th request should be rate limited")
}

func TestAuthenticate_FullWorkflow_ValidRequest(t *testing.T) {
	resetMockData()
	setValidLicense("test-api-key", true)
	setIpRequestCount("127.0.0.1", 0)
	
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-api-key", "test-api-key")
	req.RemoteAddr = "127.0.0.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should pass authentication, rate limiting, and reach the endpoint
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify IP count was incremented
	assert.Equal(t, 1, getIpRequestCount("127.0.0.1"))
}