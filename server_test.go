package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_keyValueHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		request      *http.Request
		expectedCode int
		expectedBody string
	}{
		{
			name:         "post no key parameter",
			request:      httptest.NewRequest("POST", "/", nil),
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "key is required",
		},
		{
			name:         "get no key parameter",
			request:      httptest.NewRequest("GET", "/", nil),
			expectedCode: http.StatusBadRequest,
			expectedBody: "key is required",
		},
		{
			name:         "get no existent key",
			request:      httptest.NewRequest("GET", "/nonexistentkey", nil),
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "put no key parameter",
			request:      httptest.NewRequest("PUT", "/", nil),
			expectedCode: http.StatusBadRequest,
			expectedBody: "key is required",
		},
		{
			name:         "delete no key parameter",
			request:      httptest.NewRequest("DELETE", "/", nil),
			expectedCode: http.StatusBadRequest,
			expectedBody: "key is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &keyValueHandler{}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, tt.request)
			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}
		})
	}
}

func Test_putHandler(t *testing.T) {
	h := &keyValueHandler{}
	rr := httptest.NewRecorder()
	const newkey = "newkey"
	const wantedValue = "newvalue"
	r := httptest.NewRequest("PUT", "/"+newkey, strings.NewReader(wantedValue))
	h.ServeHTTP(rr, r)
	wanted := http.StatusCreated
	status := rr.Result().StatusCode
	if status != wanted {
		t.Errorf("put handler returned wrong status code: got %v wnat %v", status, wanted)
	}
	value, _ := Get(newkey)
	if value != wantedValue {
		t.Errorf("sotored value is %v but wanted %v", value, wantedValue)
	}
}

func Test_deleteHandler(t *testing.T) {
	h := &keyValueHandler{}
	rr := httptest.NewRecorder()
	const newkey = "newkey"
	const wantedValue = "newvalue"
	r := httptest.NewRequest("DELETE", "/"+newkey, nil)
	err := Put(newkey, wantedValue)
	if err != nil {
		return
	}
	h.ServeHTTP(rr, r)
	wanted := http.StatusOK
	status := rr.Result().StatusCode
	if status != wanted {
		t.Errorf("delete handler returned wrong status code: got %v wnat %v", status, wanted)
	}
	_, e := Get(newkey)
	if !errors.Is(e, ErrorNoSuchKey) {
		t.Error("Deleted value still in store")
	}
}

func Test_getHandler(t *testing.T) {
	h := &keyValueHandler{}
	rr := httptest.NewRecorder()
	const newkey = "newkey"
	const wantedValue = "newvalue"
	r := httptest.NewRequest("GET", "/"+newkey, nil)
	err := Put(newkey, wantedValue)
	if err != nil {
		return
	}
	h.ServeHTTP(rr, r)
	wanted := http.StatusOK
	status := rr.Result().StatusCode
	if status != wanted {
		t.Errorf("get handler returned wrong status code: got %v wnat %v", status, wanted)
	}
	value, _ := io.ReadAll(rr.Result().Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(rr.Result().Body)
	if string(value) != wantedValue {
		t.Errorf("get handler returned wrong status code: got %v wnat %v", value, wantedValue)
	}
}
