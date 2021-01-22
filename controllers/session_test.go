package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"../repositories"
	"github.com/go-redis/redis"

	"github.com/alicebob/miniredis"
	"github.com/julienschmidt/httprouter"
)

const (
	testDir      = "test"
	testFile     = "test.txt"
	testFilePath = testDir + "/" + testFile
)

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

func request(body string, mux *httprouter.Router) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/sessions", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestCreateSessionOk(t *testing.T) {
	os.Mkdir(testDir, 0755)
	defer os.Remove(testDir)
	os.Create(testFilePath)
	defer os.Remove(testFilePath)

	mux, mr := createMux(testDir)
	defer mr.Close()

	b := fmt.Sprintf(`{"filename": "%s", "expectedClients": %d}`, testFile, 2)
	res := request(b, mux)

	if code := res.Code; code != http.StatusCreated {
		t.Errorf("Expected status code 201 created, got %d", code)
	}
}

func TestCreateSessionNotFound(t *testing.T) {
	os.Mkdir(testDir, 0755)
	defer os.Remove(testDir)

	mux, mr := createMux(testDir)
	defer mr.Close()

	b := fmt.Sprintf(`{"filename":"%s", "expectedClients": %d}`, testFile, 2)
	res := request(b, mux)

	if code := res.Code; code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d\n%s", code, res.Body)
	}
}

func TestCreateSessionEmptyFilename(t *testing.T) {
	os.Mkdir(testDir, 0755)
	defer os.Remove(testDir)
	os.Create(testFilePath)
	defer os.Remove(testFilePath)

	mux, mr := createMux("test")
	defer mr.Close()

	b := fmt.Sprintf(`{"filename":"", "expectedClients": %d}`, 2)
	res := request(b, mux)

	if code := res.Code; code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", code)
	}
}

func TestCreateSessionInvalidFilename(t *testing.T) {
	os.Mkdir(testDir, 0755)
	defer os.Remove(testDir)
	os.Create(testFilePath)
	defer os.Remove(testFilePath)

	mux, mr := createMux("test")
	defer mr.Close()

	b := fmt.Sprintf(`{"filename":"../main.go", "expectedClients": %d}`, 2)
	res := request(b, mux)

	if code := res.Code; code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", code)
	}
}

func createMux(d string) (*httprouter.Router, *miniredis.Miniredis) {
	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repo := repositories.NewSessionRepository(rc)

	sc := NewSessionController("test", repo)
	mux := httprouter.New()
	mux.POST("/sessions", sc.CreateSession)
	return mux, mr
}
