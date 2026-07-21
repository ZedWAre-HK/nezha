package cloudflare

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuth2Config(t *testing.T) {
	config := OAuth2Config("https://example.cloudflareaccess.com/", "client-id", "secret", "https://nezha.example.com/oauth2/callback")
	if config.Endpoint.AuthURL != "https://example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/client-id/authorization" {
		t.Fatalf("unexpected authorization URL: %s", config.Endpoint.AuthURL)
	}
	if config.Endpoint.TokenURL != "https://example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/client-id/token" {
		t.Fatalf("unexpected token URL: %s", config.Endpoint.TokenURL)
	}
	wantScopes := []string{"openid", "email", "profile"}
	if len(config.Scopes) != len(wantScopes) {
		t.Fatalf("unexpected scopes: %#v", config.Scopes)
	}
	for i, scope := range wantScopes {
		if config.Scopes[i] != scope {
			t.Fatalf("unexpected scopes: %#v", config.Scopes)
		}
	}
}

func TestFetchUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cdn-cgi/access/sso/oidc/client-id/userinfo" {
			t.Fatalf("unexpected userinfo path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"user-id","email":"Admin@Example.com","name":"Admin"}`))
	}))
	defer server.Close()

	userInfo, err := FetchUserInfo(context.Background(), server.Client(), server.URL, "client-id")
	if err != nil {
		t.Fatal(err)
	}
	if userInfo.Email != "Admin@Example.com" || userInfo.Sub != "user-id" {
		t.Fatalf("unexpected userinfo: %#v", userInfo)
	}
}

func TestFetchUserInfoRejectsErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusUnauthorized)
	}))
	defer server.Close()

	if _, err := FetchUserInfo(context.Background(), server.Client(), server.URL, "client-id"); err == nil {
		t.Fatal("expected an error for a non-2xx response")
	}
}
