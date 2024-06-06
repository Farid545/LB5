package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestForward(t *testing.T) {
    backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/some/path" {
            t.Errorf("Expected /some/path, but got %s", r.URL.Path)
        }
        rw.WriteHeader(http.StatusOK)
    }))
    defer backend.Close()

    backendURL := backend.URL

    rw := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "http://localhost/some/path", nil)

    forwardFunc := func(dst string, rw http.ResponseWriter, r *http.Request) error {
        fwdRequest := r.Clone(r.Context())
        fwdRequest.URL.Scheme = "http"
        fwdRequest.URL.Host = dst
        fwdRequest.RequestURI = ""

        resp, err := http.DefaultClient.Do(fwdRequest)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        for k, v := range resp.Header {
            rw.Header()[k] = v
        }
        rw.WriteHeader(resp.StatusCode)
        _, err = rw.Write([]byte("Forwarded"))
        return err
    }

    err := forwardFunc(backendURL[7:], rw, req)
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    if status := rw.Result().StatusCode; status != http.StatusOK {
        t.Errorf("Expected status OK, but got %d", status)
    }
}
