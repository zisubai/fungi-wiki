package auth

import "testing"

func TestTokenRoundTrip(t *testing.T) {
	service := NewTokenService("test-secret-at-least-long-enough")
	token, _, err := service.Issue(User{ID: "user-id", Email: "admin@example.com", Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-id" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestTokenRejectsDifferentSecret(t *testing.T) {
	token, _, err := NewTokenService("first-secret").Issue(User{ID: "user-id", Role: "operator"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = NewTokenService("second-secret").Parse(token); err == nil {
		t.Fatal("expected invalid token")
	}
}
