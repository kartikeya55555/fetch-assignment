package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ReceiptResponse expects the JSON: { "id": "<some-id>" }
type ReceiptResponse struct {
	ID string `json:"id"`
}

func TestIntegration_ProcessAndGetReceiptPoints(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		// We always expect 202 now (the server enqueues everything)
		expectedPostStatus int
		// This is what we want to find in the post response body (like `"id":`)
		postResponseCheck string
		// After we do a GET, we likely expect 404 because worker not done or invalid
		expectedGetStatus int
		// The substring we expect in the GET response for an unprocessed or invalid receipt
		getResponseCheck string
	}{
		{
			name: "Valid receipt (single item)",
			payload: `{
                "retailer": "Target",
                "purchaseDate": "2022-01-01",
                "purchaseTime": "13:01",
                "items": [{"shortDescription": "Mountain Dew 12PK", "price": "6.49"}],
                "total": "6.49"
            }`,
			expectedPostStatus: http.StatusAccepted, // 202
			postResponseCheck:  `"id":`,
			expectedGetStatus:  http.StatusOK,
			getResponseCheck:   `points`,
		},
		{
			name: "Valid receipt (multiple items)",
			payload: `{
                "retailer": "Walmart",
                "purchaseDate": "2022-02-02",
                "purchaseTime": "14:15",
                "items": [
                    {"shortDescription": "Bread", "price": "2.50"},
                    {"shortDescription": "Milk", "price": "3.00"}
                ],
                "total": "5.50"
            }`,
			expectedPostStatus: http.StatusAccepted, // 202
			postResponseCheck:  `"id":`,
			expectedGetStatus:  http.StatusOK,
			getResponseCheck:   `points`,
		},
		{
			name: "Missing purchase date",
			payload: `{
                "retailer": "Target",
                "purchaseTime": "13:01",
                "items": [{"shortDescription": "Mountain Dew 12PK", "price": "6.49"}],
                "total": "6.49"
            }`,
			expectedPostStatus: http.StatusAccepted,
			postResponseCheck:  `"id":`,
			// Worker would eventually reject it or never populate it
			expectedGetStatus: http.StatusBadRequest,
			getResponseCheck:  `errorMessage`,
		},
		{
			name: "Invalid JSON format",
			payload: `{
                "retailer": "Target",
                "purchaseDate": "2022-01-01",
                "purchaseTime": "13:01",
                "items": [{"shortDescription": "Mountain Dew 12PK", "price": "6.49"}
                "total": "6.49"
            }`,
			expectedPostStatus: http.StatusBadRequest,
			postResponseCheck:  `"error":`,
			expectedGetStatus:  http.StatusNotFound,
			getResponseCheck:   `receiptId doesn't exist`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1) POST /receipts/process
			postResp, postErr := http.Post(
				"http://localhost:8080/receipts/process",
				"application/json",
				bytes.NewBuffer([]byte(tc.payload)),
			)
			if postErr != nil {
				t.Fatalf("Failed to send POST request: %v", postErr)
			}
			defer postResp.Body.Close()

			assert.Equal(t, tc.expectedPostStatus, postResp.StatusCode)

			bodyBytes, _ := io.ReadAll(postResp.Body)
			bodyStr := string(bodyBytes)
			assert.Contains(t, bodyStr, tc.postResponseCheck, "POST response should contain %s", tc.postResponseCheck)

			log.Println("body str after creating **", bodyStr)
			// 2) Extract ID from the POST response
			var rr ReceiptResponse
			_ = json.Unmarshal(bodyBytes, &rr)
			if postResp.StatusCode == 404 {
				assert.NotEmpty(t, rr.ID, "Should have a valid receipt ID in the response")
			}
			log.Println(" response after creating **", rr)
			// 3) OPTIONAL: Sleep if you want to simulate time for the worker to (not) process
			time.Sleep(10 * time.Second)

			// 4) GET /receipts/{id}/points => expect some 404 or "doesn't exist" because worker isn't done or the data is invalid
			getResp, getErr := http.Get(fmt.Sprintf("http://localhost:8080/receipts/%s/points", rr.ID))
			if getErr != nil {
				t.Fatalf("Failed to send GET request: %v", getErr)
			}
			defer getResp.Body.Close()
			log.Println("response **", getResp)
			assert.Equal(t, tc.expectedGetStatus, getResp.StatusCode)

			getBodyBytes, _ := io.ReadAll(getResp.Body)
			getBodyStr := string(getBodyBytes)
			log.Println("get body str **", getBodyStr)
			assert.Contains(t, getBodyStr, tc.getResponseCheck,
				"GET response should contain %q for an unprocessed or invalid receipt", tc.getResponseCheck)
		})
	}
}

func TestIntegration_GetInvalidReceiptPoints(t *testing.T) {
	// Wait for the server to be up, if needed
	time.Sleep(2 * time.Second)

	invalidID := "non-existent-id"
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/receipts/%s/points", invalidID))
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, bodyStr, `receiptId doesn't exist`)
}
