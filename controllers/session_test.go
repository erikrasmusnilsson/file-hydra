package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestNewSessionController(t *testing.T) {
	p := "test"
	sc := NewSessionController(p)
	if sc.basePath != p {
		t.Errorf("Expected basePath to be %s, got %s.", p, sc.basePath)
	}
}

func TestIsValidFilename(t *testing.T) {
	valid := "file.txt"
	empty := ""
	emptySpace := " "
	illegal := "../file.txt"

	if !isValidFilename(valid) {
		t.Errorf("Expected isValidFilename(%s) to return true, returned false", valid)
	}

	if isValidFilename(empty) {
		t.Errorf("Expected isValidFilename(%s) to return false, returned true", empty)
	}

	if isValidFilename(emptySpace) {
		t.Errorf("Expected isValidFilename(%s) to return false, returned true", emptySpace)
	}

	if isValidFilename(illegal) {
		t.Errorf("Expected isValidFilename(%s) to return false, returned true", illegal)
	}
}

func TestCreateSessionNotFound(t *testing.T) {
	os.Mkdir("test", 0755)
	mux := createMux("test")

	bdy := `{"filename":"test-file.txt"}`
	req, _ := http.NewRequest("POST", "/sessions", strings.NewReader(bdy))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", code)
	}
	os.Remove("test")
}

func TestCreateSessionEmptyFilename(t *testing.T) {
	os.Mkdir("test", 0755)
	mux := createMux("test")

	bdy := `{"filename":""}`
	req, _ := http.NewRequest("POST", "/sessions", strings.NewReader(bdy))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", code)
	}
	os.Remove("test")
}

func TestCreateSessionInvalidFilename(t *testing.T) {
	os.Mkdir("test", 0755)
	mux := createMux("test")

	bdy := `{"filename":"../main.go"}`
	req, _ := http.NewRequest("POST", "/sessions", strings.NewReader(bdy))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if code := rr.Code; code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", code)
	}
	os.Remove("test")
}

func createMux(d string) *httprouter.Router {
	sc := NewSessionController("test")
	mux := httprouter.New()
	mux.POST("/sessions", sc.CreateSession)
	return mux
}
