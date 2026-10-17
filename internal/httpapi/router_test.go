package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(nil, nil, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf(
			"Esperado HTTP 200, recebido %d",
			recorder.Code,
		)
	}

	expected := "{\"status\":\"ok\"}\n"

	if recorder.Body.String() != expected {
		t.Errorf(
			"Resposta inesperada: %s",
			recorder.Body.String(),
		)
	}
}
