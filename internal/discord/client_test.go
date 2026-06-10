package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayload(t *testing.T) {
	update, err := Payload("hello")
	if err != nil {
		t.Fatal(err)
	}
	clear, err := Payload("")
	if err != nil {
		t.Fatal(err)
	}

	if string(update) != `{"custom_status":{"text":"hello"}}` {
		t.Fatalf("unexpected update payload: %s", update)
	}
	if string(clear) != `{"custom_status":null}` {
		t.Fatalf("unexpected clear payload: %s", clear)
	}
}

func TestClientUpdateAndClear(t *testing.T) {
	var payloads []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.Header.Get("Authorization") != "secret" {
			t.Fatal("missing authorization header")
		}

		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		payloads = append(payloads, payload)
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWithURL("secret", server.URL, server.Client())
	if err := client.Update(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if err := client.Clear(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(payloads) != 2 {
		t.Fatalf("expected two payloads, got %d", len(payloads))
	}
	if payloads[1]["custom_status"] != nil {
		t.Fatalf("expected null clear payload, got %#v", payloads[1])
	}
}

func TestClientErrorDoesNotExposeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewWithURL("sensitive-token", server.URL, server.Client())
	err := client.Update(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected request error")
	}
	if contains := stringContains(err.Error(), "sensitive-token"); contains {
		t.Fatal("error exposed token")
	}
}

func stringContains(text, value string) bool {
	for index := 0; index+len(value) <= len(text); index++ {
		if text[index:index+len(value)] == value {
			return true
		}
	}
	return false
}
