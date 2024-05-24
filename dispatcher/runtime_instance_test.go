package dispatcher

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeInstance_Invoke_Success(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading request body: %v", err)
		}
		expectedBody := `{"prompt": "test prompt"}`
		if string(body) != expectedBody {
			t.Errorf("Expected request body %s, got %s", expectedBody, body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response": "test response"}`))
	}))
	defer mockServer.Close()

	// Create an instance of RuntimeInstance with the mock server URL
	instance := RuntimeInstance{
		ID:  "test-instance",
		Url: mockServer.URL,
	}

	// Invoke the method
	response, err := instance.Invoke("test prompt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedResponse := `{"response": "test response"}`
	if response != expectedResponse {
		t.Errorf("Expected response %s, got %s", expectedResponse, response)
	}
}

func TestRuntimeInstance_Invoke_Error(t *testing.T) {
	// Create an instance of RuntimeInstance with an invalid URL
	instance := RuntimeInstance{
		ID:  "test-instance",
		Url: "http://invalid-url",
	}

	// Invoke the method
	_, err := instance.Invoke("test prompt")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
