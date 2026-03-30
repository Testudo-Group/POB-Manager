package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenClaims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"type"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

type TokenManager struct {
	secret           []byte
	accessTTLMinutes int
	refreshTTLHours  int
}

func NewTokenManager(secret string, accessTTLMinutes, refreshTTLHours int) *TokenManager {
	return &TokenManager{
		secret:           []byte(secret),
		accessTTLMinutes: accessTTLMinutes,
		refreshTTLHours:  refreshTTLHours,
	}
}

func (m *TokenManager) GenerateAccessToken(subject, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(m.accessTTLMinutes) * time.Minute)
	token, err := m.generate(subject, email, role, AccessTokenType, expiresAt)
	return token, expiresAt, err
}

func (m *TokenManager) GenerateRefreshToken(subject, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(m.refreshTTLHours) * time.Hour)
	token, err := m.generate(subject, email, role, RefreshTokenType, expiresAt)
	return token, expiresAt, err
}

func (m *TokenManager) Validate(token, expectedType string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := m.sign(unsigned)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (m *TokenManager) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *TokenManager) generate(subject, email, role, tokenType string, expiresAt time.Time) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claims := TokenClaims{
		Subject:   subject,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  time.Now().UTC().Unix(),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := m.sign(unsigned)
	return unsigned + "." + signature, nil
}

func (m *TokenManager) sign(unsigned string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
