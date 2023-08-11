package router

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {

	router := NewRouter()

	// Create a test HTTP server
	server := httptest.NewServer(router.Mux)
	defer server.Close()

	t.Run("AddRoute", func(t *testing.T) {
		// Set up a test route
		router.AddRoute("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Test Response"))
		})

		// Perform a request to the test route
		resp, err := http.Get(server.URL + "/test")
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Check if the response status code is 200
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check if the response body is correct
		expectedBody := "Test Response"
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, string(body))
	})
}
