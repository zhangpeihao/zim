// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

type TestHTTPServer struct {
}

func (srv *TestHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func TestHTTPSListener(t *testing.T) {
	listener, err := NewHTTPSListener("./httpcert/cert.pem", "./httpcert/key.pem", ":12341")
	if err != nil {
		t.Fatalf("NewHTTPSListener error: %s\n", err)
	}

	srv := &TestHTTPServer{}
	httpSrv := &http.Server{Handler: srv}

	var httpsErr error
	go func() {
		httpsErr = httpSrv.Serve(listener)
	}()

	time.Sleep(time.Second)
	if httpsErr != nil {
		t.Fatalf("HTTP serve error: %s\n", httpsErr)
	}

	insecureHTTPSClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := insecureHTTPSClient.Get("https://localhost:12341/test")
	if err != nil {
		t.Fatal("http.Get error:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("http.Get response status code: %d\n", resp.StatusCode)
	}
}

func TestError(t *testing.T) {
	_, err := NewHTTPSListener("./httpcert/nofile.pem", "./httpcert/nofile.pem", ":12341")
	if err == nil {
		t.Error("NewHTTPSListener should return error with unexisted file path\n")
	}
	_, err = NewHTTPSListener("./httpcert/cert.pem", "./httpcert/key.pem", "error address")
	if err == nil {
		t.Error("NewHTTPSListener should return error with error address\n")
	}
}