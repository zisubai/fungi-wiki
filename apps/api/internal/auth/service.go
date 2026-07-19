package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type TokenService struct {
	secret   []byte
	duration time.Duration
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{[]byte(secret), 12 * time.Hour}
}
func (s *TokenService) Issue(u User) (string, time.Time, error) {
	expires := time.Now().Add(s.duration)
	claims := Claims{Role: u.Role, Email: u.Email, RegisteredClaims: jwt.RegisteredClaims{Subject: u.ID, ExpiresAt: jwt.NewNumericDate(expires), IssuedAt: jwt.NewNumericDate(time.Now()), Issuer: "fungi-wiki"}}
	token, e := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	return token, expires, e
}
func (s *TokenService) Parse(value string) (*Claims, error) {
	token, e := jwt.ParseWithClaims(value, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return s.secret, nil
	})
	if e != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
