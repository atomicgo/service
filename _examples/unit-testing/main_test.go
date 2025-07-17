package main

import (
	"io"
	"net/http"
	"testing"
)

func TestHello(t *testing.T) {
	t.Parallel()

	ts := getService().TestServer()
	defer ts.Close()

	res, err := http.Get(ts.URL + "/hello/world")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello, world!" {
		t.Fatalf("expected body 'Hello, world!', got %s", string(body))
	}
}
