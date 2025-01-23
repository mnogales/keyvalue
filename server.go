package main

import (
	"errors"
	"io"
	"net/http"
	"regexp"
)

type keyValueHandler struct{}

func (h *keyValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet:
		h.GetHandler(w, r)
		return
	case r.Method == http.MethodDelete:
		h.DeleteHandler(w, r)
		return
	case r.Method == http.MethodPut:
		h.PutHandler(w, r)
		return
	default:
		methodNotAllowed(w)
		return
	}
}

func (h *keyValueHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key, err := getKey(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	value, err := Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalServerError(w, err)
		return
	}

	_, e := w.Write([]byte(value))
	if e != nil {
		internalServerError(w, err)
		return
	}
}

func (h *keyValueHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key, err := getKey(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	err = Delete(key)
	if err != nil {
		internalServerError(w, err)
		return
	}
}

func (h *keyValueHandler) PutHandler(w http.ResponseWriter, r *http.Request) {
	key, err := getKey(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	value, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)

	if err != nil {
		internalServerError(w, err)
		return
	}

	err = Put(key, string(value))
	if err != nil {
		internalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getKey(r *http.Request) (string, error) {
	matches := regexp.MustCompile(`^/(.+)$`).FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		return "", errors.New("key param missing")
	}
	return matches[1], nil
}

func methodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func internalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, e := w.Write([]byte(err.Error()))
	if e != nil {
		return
	}
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, e := w.Write([]byte(err.Error()))
	if e != nil {
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", &keyValueHandler{})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}
